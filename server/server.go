package server

import (
	"github.com/google/uuid"
	"github.com/ragoob/gCache/db"
	"github.com/ragoob/gCache/pkg/client"
	"sync"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts
	NodeID    string
	followers map[*client.Client]struct{}
	db        db.DB
	mtx       sync.Mutex
}

func generateNodeID() string {
	return uuid.New().String()
}

func NewServer(opts ServerOpts, db db.DB) *Server {
	return &Server{
		NodeID:     generateNodeID(),
		ServerOpts: opts,
		db:         db,
		followers:  make(map[*client.Client]struct{}),
	}
}
