## simple go in-memory Distributed key-value store

### Features
- [x] Read key
- [x] Write key
- [x] Leader - followers
- [ ] Leader election (uder development using Raft and Blot store)
- [ ] Disaster recovery under development using write a head log)

#### Build 
```
go build .
```
#### Run as a leader 
```
./gCache  --listenaddr :8000 

```

### Run as follower node 

```
./gCache  --listenaddr :5000 --leaderaddr :8000
```
