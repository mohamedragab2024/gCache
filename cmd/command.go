package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Status byte

func (s Status) String() string {
	switch s {
	case Error:
		return "ERR"
	case OK:
		return "OK"
	case NotExists:
		return "KEYNOTFOUND"
	case LeaderError:
		return "INVALID LEADER"
	default:
		return "NONE"
	}
}

const (
	None Status = iota
	OK
	Error
	NotExists
	LeaderError
)

type Command byte

const (
	Empty Command = iota
	Set
	Get
	Join
)

type SetRes struct {
	Status Status
}

type GetRes struct {
	Status Status
	Val    []byte
}
type JoinRes struct {
	Addr []byte
}

type JoinCmd struct {
	Addr []byte
}

type SetCmd struct {
	Key         []byte
	Val         []byte
	Replication bool
	Duration    int
}

type GetCmd struct {
	Key []byte
}

func (r *GetRes) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Status)
	len := int32(len(r.Val))
	binary.Write(buf, binary.LittleEndian, len)
	binary.Write(buf, binary.LittleEndian, r.Val)
	return buf.Bytes()
}

func (r *SetRes) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Status)
	return buf.Bytes()
}

func (c *JoinCmd) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, Join)
	len := int32(len(c.Addr))
	binary.Write(buf, binary.LittleEndian, len)
	binary.Write(buf, binary.LittleEndian, c.Addr)
	return buf.Bytes()
}

func (c *GetCmd) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, Get)
	len := int32(len(c.Key))
	binary.Write(buf, binary.LittleEndian, len)
	binary.Write(buf, binary.LittleEndian, c.Key)
	return buf.Bytes()
}

func (c *SetCmd) GetBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, Set)
	binary.Write(buf, binary.LittleEndian, int32(len(c.Key)))
	binary.Write(buf, binary.LittleEndian, c.Key)
	binary.Write(buf, binary.LittleEndian, int32(len(c.Val)))
	binary.Write(buf, binary.LittleEndian, c.Val)
	binary.Write(buf, binary.LittleEndian, int32(c.Duration))
	binary.Write(buf, binary.LittleEndian, c.Duration)
	return buf.Bytes()
}

func ParseGetRes(r io.Reader) (*GetRes, error) {
	resp := &GetRes{}
	binary.Read(r, binary.LittleEndian, &resp.Status)
	var len int32
	binary.Read(r, binary.LittleEndian, &len)
	resp.Val = make([]byte, len)
	binary.Read(r, binary.LittleEndian, &resp.Val)
	return resp, nil
}

func ParseSetRes(r io.Reader) (*SetRes, error) {
	resp := &SetRes{}
	err := binary.Read(r, binary.LittleEndian, &resp.Status)
	return resp, err
}

func ParseJoinRes(r io.Reader) (*JoinRes, error) {
	resp := &JoinRes{}
	binary.Read(r, binary.LittleEndian, &resp.Addr)
	return resp, nil
}

func ParseCmd(r io.Reader) (any, error) {
	var cmd Command
	if err := binary.Read(r, binary.LittleEndian, &cmd); err != nil {
		return nil, err
	}
	switch cmd {
	case Set:
		return parseSetCommand(r), nil
	case Get:
		return parseGetCommand(r), nil
	case Join:
		return parseJoinCommand(r), nil
	default:
		return nil, fmt.Errorf("invalid command")
	}
}

func parseSetCommand(r io.Reader) *SetCmd {
	cmd := &SetCmd{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	var valLen int32
	binary.Read(r, binary.LittleEndian, &valLen)
	cmd.Val = make([]byte, valLen)
	binary.Read(r, binary.LittleEndian, &cmd.Val)

	var ttl int32
	binary.Read(r, binary.LittleEndian, &ttl)
	cmd.Duration = int(ttl)
	var replication bool
	binary.Read(r, binary.LittleEndian, &replication)
	cmd.Replication = replication
	return cmd
}

func parseGetCommand(r io.Reader) *GetCmd {
	cmd := &GetCmd{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	return cmd
}

func parseJoinCommand(r io.Reader) *JoinCmd {
	cmd := &JoinCmd{}
	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Addr = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Addr)
	return cmd
}

func ParseSetCommand(r io.Reader) (*SetCmd, error) {
	var cmd Command
	if err := binary.Read(r, binary.LittleEndian, &cmd); err != nil {
		return nil, err
	}
	switch cmd {
	case Set:
		return parseSetCommand(r), nil
	default:
		return nil, fmt.Errorf("invalid command")
	}
}
