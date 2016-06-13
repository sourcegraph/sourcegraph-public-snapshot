package gitserver

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/sourcegraph/chanrpc"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type request struct {
	Exec   *execRequest
	Search *searchRequest
	Create *createRequest
	Remove *removeRequest
}

var ReposDir string
var servers [](chan<- *request)

func Serve(l net.Listener) error {
	registerMetrics()
	requests := make(chan *request, 100)
	go processRequests(requests)
	srv := &chanrpc.Server{RequestChan: requests}
	return srv.Serve(l)
}

func processRequests(requests <-chan *request) {
	for req := range requests {
		if req.Exec != nil {
			go handleExecRequest(req.Exec)
		}
		if req.Search != nil {
			go handleSearchRequest(req.Search)
		}
		if req.Create != nil {
			go handleCreateRequest(req.Create)
		}
		if req.Remove != nil {
			go handleRemoveRequest(req.Remove)
		}
	}
}

func Connect(addr string) {
	requestsChan := make(chan *request, 100)
	servers = append(servers, requestsChan)

	go func() {
		for {
			err := chanrpc.DialAndDeliver(addr, requestsChan)
			log.Printf("gitserver: DialAndDeliver error: %v", err)
			time.Sleep(time.Second)
		}
	}()
}

type genericReply interface {
	repoFound() bool
}

var errRPCFailed = errors.New("gitserver: rpc failed")

func broadcastCall(newRequest func() (*request, func() (genericReply, bool))) (interface{}, error) {
	allReplies := make(chan genericReply, len(servers))
	for _, server := range servers {
		req, getReply := newRequest()
		server <- req
		go func() {
			reply, ok := getReply()
			if !ok {
				allReplies <- nil
				return
			}
			allReplies <- reply
		}()
	}

	rpcError := false
	for range servers {
		reply := <-allReplies
		if reply == nil {
			rpcError = true
			continue
		}
		if reply.repoFound() {
			return reply, nil
		}
	}
	if rpcError {
		return nil, errRPCFailed
	}
	return nil, vcs.RepoNotExistError{}
}
