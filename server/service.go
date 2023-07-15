package server

import (
	"errors"
	"fmt"
	"github.com/ragoob/gCache/cmd"
	"github.com/ragoob/gCache/db"
	"github.com/ragoob/gCache/pkg/client"
	pb "github.com/ragoob/gCache/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	pb.GCacheServiceServer
	ServerOpts
	followers map[*client.Client]struct{}
	db        db.DB
	mu        sync.Mutex
}

func NewServer(opts ServerOpts, db db.DB) *Server {
	return &Server{
		ServerOpts: opts,
		db:         db,
		followers:  make(map[*client.Client]struct{}),
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: [%v]", err)
	}

	server := grpc.NewServer()
	pb.RegisterGCacheServiceServer(server, s)

	log.Printf("Server started [%s]", s.ListenAddr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := s.db.Get(req.Key)
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{
		Value: val,
	}, nil
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	command := &cmd.SetCmd{
		Key:      req.Key,
		Val:      req.Value,
		Duration: 5,
	}

	// check if not leader don't write
	if !s.IsLeader && !command.Replication {
		return &pb.SetResponse{Success: false}, errors.New("replica can't write")
	}
	// set the value to db
	if err := s.db.Set(command.Key, command.Val, time.Duration(command.Duration)); err != nil {
		return &pb.SetResponse{Success: false}, err
	}
	// set the value to the followers replicas
	go func() {
		for peer := range s.followers {
			if err := peer.Replicate(context.TODO(), command.Key, command.Val, command.Duration); err != nil {
				log.Println("replicating to follower error:", err)
			}
		}
	}()
	// complete set value
	return &pb.SetResponse{Success: true}, nil
}
