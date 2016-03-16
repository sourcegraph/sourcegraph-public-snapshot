package gitserver

import (
	"log"
	"net/rpc"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type Git struct {
}

var ReposDir string
var servers [](chan<- *rpc.Call)

func RegisterHandler() {
	rpc.Register(&Git{})
	rpc.HandleHTTP()
}

func Dial(addr string) error {
	clientSingleton, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return err
	}

	callChan := make(chan *rpc.Call, 10)
	servers = append(servers, callChan)
	resetConnectionChan := make(chan *rpc.Client)

	go func() {
		for {
			select {
			case call := <-callChan:
				clientForCall := clientSingleton
				done := make(chan *rpc.Call, 1)
				clientForCall.Go(call.ServiceMethod, call.Args, call.Reply, done)
				go func() {
					call.Error = (<-done).Error
					if call.Error == rpc.ErrShutdown {
						resetConnectionChan <- clientForCall
						callChan <- call // retry
						return
					}
					call.Done <- call
				}()

			case client := <-resetConnectionChan:
				if client != clientSingleton {
					continue
				}
				clientSingleton.Close()
				for {
					newClient, err := rpc.DialHTTP("tcp", addr)
					if err != nil {
						log.Printf("dial to git server failed: %s", err)
						time.Sleep(time.Second)
						continue
					}
					clientSingleton = newClient
					break
				}
			}
		}
	}()

	return nil
}

type repoExistsReply interface {
	repoExists() bool
}

func broadcastCall(serviceMethod string, args interface{}, newReply func() repoExistsReply) (interface{}, error) {
	done := make(chan *rpc.Call, len(servers))
	for _, server := range servers {
		server <- &rpc.Call{
			ServiceMethod: serviceMethod,
			Args:          args,
			Reply:         newReply(),
			Done:          done,
		}
	}
	var rpcError error
	for range servers {
		call := <-done
		if call.Error != nil {
			rpcError = call.Error
			continue
		}
		if call.Reply.(repoExistsReply).repoExists() {
			return call.Reply, nil
		}
	}
	if rpcError != nil {
		return nil, rpcError
	}
	return nil, vcs.ErrRepoNotExist
}
