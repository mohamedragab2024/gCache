package new

import (
	"context"
	"github.com/ragoob/gCache/cmd"
	"github.com/ragoob/gCache/pkg/client"
	"io"
	"log"
	"net"
	"time"
)

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
	if !s.IsLeader && !command.Replication {
		resp.Status = cmd.Error
		_, err := conn.Write(resp.GetBytes())
		return err
	}
	if err := s.db.Set(command.Key, command.Val, time.Duration(command.Duration)); err != nil {
		resp.Status = cmd.Error
		_, err := conn.Write(resp.GetBytes())
		return err
	}
	resp.Status = cmd.OK
	_, err := conn.Write(resp.GetBytes())
	go func() {
		for peer := range s.followers {
			if err := peer.Replicate(context.TODO(), command.Key, command.Val, command.Duration); err != nil {
				log.Println("replicating to follower error:", err)
			}
		}
	}()
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
	c, err := client.Connect(string(command.Addr), client.Options{})
	if err != nil {
		log.Printf("error join cluster [%s]", string(command.Addr))
		return nil
	}
	s.followers[c] = struct{}{}
	return nil
}
