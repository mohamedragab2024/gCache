package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ragoob/gCache/cmd"
	"io"
	"log"
	"time"
)

// Apply applies a Raft log entry to the dummy FSM
func (s *Server) Apply(l *raft.Log) interface{} {

	// Extract the command data from the log entry
	reader := bytes.NewReader(l.Data)
	// Deserialize the command
	command, err := cmd.ParseSetCommand(reader)
	if err != nil {
		return err
	}

	if err := s.db.Set(command.Key, command.Val, time.Duration(command.Duration)); err != nil {
		return err
	}

	return nil
}

// SetCmd represents a set command for a key-value pair
type SetCmd struct {
	Key         []byte
	Val         []byte
	Replication bool
	Duration    int
}

// ServerSnapshot represents a snapshot of the cache server's state
type ServerSnapshot struct {
	Commands []SetCmd
}

// Snapshot writes a snapshot of the cache server's state
func (s *Server) Snapshot() (raft.FSMSnapshot, error) {
	// Create a new snapshot
	snapshot := &ServerSnapshot{}

	// Acquire a read lock on the cache to ensure consistent snapshot
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Iterate over the cache and collect the set commands
	for key, val := range s.db.Store() {
		cmd := SetCmd{
			Key:         []byte(key),
			Val:         val,
			Replication: false, // Set the replication flag as needed
			Duration:    0,     // Set the duration as needed
		}
		snapshot.Commands = append(snapshot.Commands, cmd)
	}

	log.Println("Snapshot created successfully")

	return snapshot, nil
}

// Persist persists the snapshot data
func (s *ServerSnapshot) Persist(sink raft.SnapshotSink) error {
	// Serialize the snapshot to a buffer
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(s)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("failed to encode snapshot: %v", err)
	}

	// Write the serialized snapshot data to the sink
	_, err = sink.Write(buf.Bytes())
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("failed to write snapshot: %v", err)
	}

	// Close the sink to finalize the snapshot
	err = sink.Close()
	if err != nil {
		return fmt.Errorf("failed to close snapshot sink: %v", err)
	}

	log.Println("Snapshot persisted successfully")

	return nil
}

// Release releases any resources associated with the snapshot
func (s *ServerSnapshot) Release() {
	// Perform any necessary cleanup or release of resources
}

// Restore restores the cache server's state from a snapshot
func (s *Server) Restore(snapshot io.ReadCloser) error {
	// Read the snapshot data from the provided io.ReadCloser
	data, err := io.ReadAll(snapshot)
	if err != nil {
		return fmt.Errorf("failed to read snapshot data: %v", err)
	}

	// Create a new buffer to decode the snapshot data
	buf := bytes.NewBuffer(data)

	// Decode the snapshot data into a new ServerSnapshot instance
	var restoredSnapshot ServerSnapshot
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&restoredSnapshot)
	if err != nil {
		return fmt.Errorf("failed to decode snapshot data: %v", err)
	}

	// Acquire a write lock on the cache to restore the state
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Clear the existing cache
	s.db.Clear()

	// Restore the set commands from the snapshot
	for _, cmd := range restoredSnapshot.Commands {
		err := s.db.Set(cmd.Key, cmd.Val, time.Duration(cmd.Duration))
		if err != nil {
			return fmt.Errorf("failed to restore command: %v", err)
		}
	}

	log.Println("Snapshot restored successfully")

	return nil
}
