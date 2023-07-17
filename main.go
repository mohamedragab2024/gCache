package main

import (
	"flag"
	"log"

	"github.com/ragoob/gCache/db"
	"github.com/ragoob/gCache/server"
)

func main() {
	var (
		listenAddr = flag.String("listenaddr", ":3000", "listen address of the server")
		leaderAddr = flag.String("leaderaddr", "", "listen address of the leader")
	)
	flag.Parse()
	opts := server.ServerOpts{
		ListenAddr: *listenAddr,
		IsLeader:   len(*leaderAddr) == 0,
		LeaderAddr: *leaderAddr,
	}

	s := server.NewServer(opts, db.New())

	if err := s.Serve(); err != nil {
		log.Fatalf("Serve Error : %+v", err)
	}

}
