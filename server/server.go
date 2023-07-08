package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ragoob/gCache/client"
	"github.com/ragoob/gCache/cmd"
	"github.com/ragoob/gCache/db"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts
	followers map[*client.Client]struct{}
	db        db.DB
	mu        *sync.Mutex
}

func NewServer(opts ServerOpts, db db.DB) *Server {
	return &Server{
		ServerOpts: opts,
		db:         db,
		followers:  make(map[*client.Client]struct{}),
	}
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("Listen error: [%v]", err)
	}

	if !s.IsLeader && s.ListenAddr != "" {
		go func() {
			if err := s.dailLeader(); err != nil {
				log.Println(err)
			}
		}()
	}

	log.Printf("Server started [%s]", s.ListenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("connection error [%v]", err)
		}

		go s.handleConn(conn)
	}
}

func (s *Server) dailLeader() error {
	conn, err := net.Dial("tcp", s.LeaderAddr)
	if err != nil {
		return fmt.Errorf("Failed to connect to leader [%v]", err)
	}

	log.Println("Connected to leader")

	binary.Write(conn, binary.LittleEndian, cmd.Join)

	s.handleConn(conn)

	return nil
}
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		command, err := cmd.ParseCmd(conn)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Println("command not defiend", err)

			break
		}

		go s.handleCommand(conn, command)
	}
}

func (s *Server) handleCommand(conn net.Conn, command any) {
	switch c := command.(type) {
	case *cmd.SetCmd:
		s.handleSetCommand(conn, c)
	case *cmd.GetCmd:
		s.handleGetCommand(conn, c)
	case *cmd.JoinCmd:
		s.handleJoinCommand(conn, c)
	}
}

func (s *Server) handleSetCommand(conn net.Conn, command *cmd.SetCmd) error {
	resp := cmd.SetRes{}
	if err := s.db.Set(command.Key, command.Val, time.Duration(command.Duration)); err != nil {
		resp.Status = cmd.Error
		_, err := conn.Write(resp.GetBytes())
		return err
	}
	resp.Status = cmd.OK
	_, err := conn.Write(resp.GetBytes())
	if err == nil {
		go s.Replicate(command)
	}
	return err
}

func (s *Server) handleGetCommand(conn net.Conn, command *cmd.GetCmd) error {
	resp := cmd.GetRes{}

	val, err := s.db.Get(command.Key)
	if err != nil {
		resp.Status = cmd.Error
		_, err := conn.Write(resp.GetBytes())
		return err
	}
	resp.Status = cmd.OK
	resp.Val = val

	_, err = conn.Write(resp.GetBytes())
	return err

}

func (s *Server) handleJoinCommand(conn net.Conn, command *cmd.JoinCmd) error {
	log.Println("New follower joined: ", conn.RemoteAddr())
	s.mu.Lock()
	defer s.mu.Unlock()
	s.followers[client.New(conn)] = struct{}{}
	return nil
}

func (s *Server) Replicate(command *cmd.SetCmd) {
	for f := range s.followers {
		if err := f.Set(context.Background(), command.Key, command.Val, command.Duration); err != nil {
			log.Printf("Error replicating to follower [%d]", err)
		}
	}
}
