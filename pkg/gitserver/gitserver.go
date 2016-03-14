package gitserver

import (
	"log"
	"net/rpc"
	"time"
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
