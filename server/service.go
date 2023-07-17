package server

import (
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/rafterrors"
	"github.com/ragoob/gCache/cmd"
	pb "github.com/ragoob/gCache/proto"
	"golang.org/x/net/context"
	"time"
)

func (r rpcInterface) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	command := &cmd.SetCmd{
		Key:      req.Key,
		Val:      req.Value,
		Duration: 5,
	}
	cmdBytes := command.GetBytes()

	f := r.raft.Apply(cmdBytes, time.Second)
	if err := f.Error(); err != nil {
		return nil, rafterrors.MarkRetriable(err)
	}
	fmt.Printf("%+v :: CommitIndex: %+v \n", r.server.NodeID, f.Index())

	// complete set value
	return &pb.SetResponse{Success: true}, nil

	//// check if not leader don't write
	// todo :: leader only can write
	//if !s.IsLeader && !command.Replication {
	//	return &pb.SetResponse{Success: false}, errors.New("replica can't write")
	//}
	//// set the value to db
	//if err := s.db.Set(command.Key, command.Val, time.Duration(command.Duration)); err != nil {
	//	return &pb.SetResponse{Success: false}, err
	//}
	//// set the value to the followers replicas
	//go func() {
	//	for peer := range s.followers {
	//		if err := peer.Replicate(context.TODO(), command.Key, command.Val, command.Duration); err != nil {
	//			log.Println("replicating to follower error:", err)
	//		}
	//	}
	//}()
	//// complete set value
	//return &pb.SetResponse{Success: true}, nil
}

func (r rpcInterface) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := r.server.db.Get(req.Key)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s ::get command rpc : %s \n", r.server.NodeID, val)

	return &pb.GetResponse{
		Value: val,
	}, nil
}
