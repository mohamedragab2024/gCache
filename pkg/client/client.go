package client

import (
	"context"
	"fmt"
	"net"

	"github.com/ragoob/gCache/cmd"
)

type Options struct {
}
type Client struct {
	Options
	conn         net.Conn
	IsReplicator bool
}

func New(conn net.Conn) *Client {
	return &Client{
		conn:         conn,
		IsReplicator: true,
	}
}

func Connect(host string, opts Options) (*Client, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	return &Client{
		Options: opts,
		conn:    conn,
	}, nil
}

func (c *Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	fmt.Printf("get  key [%s] \n", string(key))
	command := &cmd.GetCmd{
		Key: key,
	}

	_, err := c.conn.Write(command.GetBytes())
	if err != nil {
		return nil, err
	}

	resp, err := cmd.ParseGetRes(c.conn)

	if err != nil {
		return nil, err
	}

	if resp.Status == cmd.NotExists {
		return nil, fmt.Errorf("key [%s] does not exsist", key)
	}

	if resp.Status != cmd.OK {
		return nil, fmt.Errorf("failed to get key [%s]", key)
	}

	return resp.Val, nil
}

func (c *Client) Set(ctx context.Context, key []byte, val []byte, duration int) error {
	fmt.Printf("Set key [%s] \n", string(key))
	command := &cmd.SetCmd{
		Key:      key,
		Val:      val,
		Duration: duration,
	}
	_, err := c.conn.Write(command.GetBytes())

	if err != nil {
		return err
	}

	resp, err := cmd.ParseSetRes(c.conn)
	if err != nil {
		return err
	}
	if resp.Status != cmd.OK {
		return fmt.Errorf("failed to write key [%s]", key)
	}

	return nil
}

func (c *Client) Replicate(ctx context.Context, key []byte, val []byte, duration int) error {
	fmt.Printf("Set key [%s] \n", string(key))
	command := &cmd.SetCmd{
		Key:         key,
		Val:         val,
		Duration:    duration,
		Replication: true,
	}
	_, err := c.conn.Write(command.GetBytes())

	if err != nil {
		return err
	}

	resp, err := cmd.ParseSetRes(c.conn)
	if err != nil {
		return err
	}
	if resp.Status != cmd.OK {
		return fmt.Errorf("failed to write key [%s]", key)
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
