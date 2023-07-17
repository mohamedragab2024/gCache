package client

import (
	"context"
	"fmt"
	pb "github.com/ragoob/gCache/proto"
	"google.golang.org/grpc"
)

type Options struct {
}
type Client struct {
	Grpc pb.GCacheServiceClient
	Options
	conn         *grpc.ClientConn
	IsReplicator bool
}

func Connect(host string, opts Options) (*Client, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewGCacheServiceClient(conn)

	return &Client{
		Grpc:    client,
		Options: opts,
		conn:    conn,
	}, nil
}

func (c *Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	fmt.Printf("get  key [%s] \n", string(key))

	res, err := c.Grpc.Get(ctx, &pb.GetRequest{Key: key})

	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func (c *Client) Set(ctx context.Context, key []byte, val []byte, duration int) error {
	fmt.Printf("Set key [%s] \n", string(key))

	_, err := c.Grpc.Set(ctx, &pb.SetRequest{Key: key, Value: val})

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
