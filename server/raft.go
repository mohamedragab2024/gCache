package server

import (
	"errors"
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	pb "github.com/ragoob/gCache/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type rpcInterface struct {
	pb.GCacheServiceServer
	server *Server
	raft   *raft.Raft
}

func (s *Server) NewRaft(ctx context.Context, myID, myAddress string, fsm raft.FSM) (*raft.Raft, *transport.Manager, error) {
	// Create a new Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(myID)

	// Create the Raft log store (replace with your own implementation)
	logStore := raft.NewInmemStore()

	// Create the Raft stable store (replace with your own implementation)
	stableStore := raft.NewInmemStore()

	baseDir := filepath.Join("raft_data/", myID)

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	// Create the transport layer (replace with your own implementation)
	tm := transport.New(raft.ServerAddress(myAddress), []grpc.DialOption{grpc.WithInsecure()})

	//timeout := time.Second * 5
	//tr, err := raft.NewTCPTransport(myAddress, nil, 10, timeout, os.Stdout)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("failed to create transport: %v", err)
	//}

	// Create the Raft node
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, fss, tm.Transport())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Raft node: %v", err)
	}
	s1 := raft.Server{
		Suffrage: raft.Voter,
		ID:       config.LocalID,
		Address:  raft.ServerAddress("127.0.0.1:4000"),
	}
	//
	//s2 := raft.Server{
	//	Suffrage: raft.Voter,
	//	ID:       raft.ServerID("node2"),
	//	Address:  raft.ServerAddress("127.0.0.1:5002"),
	//}
	//
	//s3 := raft.Server{
	//	Suffrage: raft.Voter,
	//	ID:       raft.ServerID("node3"),
	//	Address:  raft.ServerAddress("127.0.0.1:5003"),
	//}

	serverConfig := raft.Configuration{
		Servers: []raft.Server{s1},
	}

	if err := r.BootstrapCluster(serverConfig).Error(); err != nil {
		log.Fatalf("failed to bootstrab the cluster : %+v ", err)
	}

	return r, tm, nil
}

func (s *Server) Serve() error {
	// Create the Raft node
	ctx := context.Background()

	bindAddr := fmt.Sprintf("127.0.0.1%s", s.ListenAddr)

	r, tm, err := s.NewRaft(ctx, s.NodeID, bindAddr, s)
	if err != nil {
		return err
	}

	// Create the gRPC server
	server := grpc.NewServer()

	// Register the Raft node as a gRPC service
	//raftgrpc.RegisterRaftServer(server, raftNode)

	// Register your gRPC service with the server
	pb.RegisterGCacheServiceServer(server, &rpcInterface{
		server: s,
		raft:   r,
	})

	// Register the Raft node's gRPC service with the Time Machine
	tm.Register(server)
	leaderhealth.Setup(r, server, []string{"Example"})
	raftadmin.Register(server, r)
	reflection.Register(server)

	// Create a listener for the gRPC server
	lis, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}

	// Start serving gRPC requests in a separate goroutine
	go func() {
		log.Printf("Server started [%s]", s.ListenAddr)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for the Raft node to become the leader
	select {
	case <-r.LeaderCh():
		log.Println("Raft node elected as leader")
	case <-time.After(5 * time.Second):
		return errors.New("timeout: Raft node did not become leader")
	}

	// Wait for server shutdown
	<-ctx.Done()

	// Stop the gRPC server
	server.Stop()

	// Stop the Raft node
	r.Shutdown()

	return nil
}
