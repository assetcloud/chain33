// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"net"
	"time"

	"github.com/33cn/chain33/client"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
)

var (
	listenAddr            = "localhost:8805"
	unSyncMaxTimes uint32 = 6 //max 6 times
	checkInterval  uint32 = 5 // 5s
)

// HealthCheckServer  a node's health check server
type HealthCheckServer struct {
	api  client.QueueProtocolAPI
	l    net.Listener
	quit chan struct{}
}

// Close NewHealthCheckServer close
func (s *HealthCheckServer) Close() {
	close(s.quit)
}

// NewHealthCheckServer new json rpcserver object
func NewHealthCheckServer(c queue.Client) *HealthCheckServer {
	h := &HealthCheckServer{}
	h.api, _ = client.New(c, nil)
	return h
}

// Start HealthCheckServer start
func (s *HealthCheckServer) Start(cfg *types.HealthCheck) {
	if cfg != nil {
		if cfg.ListenAddr != "" {
			listenAddr = cfg.ListenAddr
		}
		if cfg.CheckInterval != 0 {
			checkInterval = cfg.CheckInterval
		}
		if cfg.UnSyncMaxTimes != 0 {
			unSyncMaxTimes = cfg.UnSyncMaxTimes
		}
	}
	log.Info("healthCheck start ", "addr", listenAddr, "inter", checkInterval, "times", unSyncMaxTimes)
	s.quit = make(chan struct{})
	go s.healthCheck()

}

func (s *HealthCheckServer) listen(on bool) error {
	if on {
		listener, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}
		s.l = listener
		log.Info("healthCheck listen open")
		return nil
	}

	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			return err
		}
		log.Info("healthCheck listen close")
		s.l = nil
	}

	return nil
}

func (s *HealthCheckServer) getHealth(sync bool) (bool, error) {
	reply, err := s.api.IsSync()
	if err != nil {
		return false, err
	}

	peerList, err := s.api.PeerInfo()
	if err != nil {
		return false, err
	}

	log.Info("healthCheck tick", "peers", len(peerList.Peers), "isSync", reply.IsOk, "sync", sync)

	return len(peerList.Peers) > 1 && reply.IsOk, nil
}

func (s *HealthCheckServer) healthCheck() {
	ticker := time.NewTicker(time.Second * time.Duration(checkInterval))
	defer ticker.Stop()

	var sync bool
	var unSyncTimes uint32

	for {
		select {
		case <-s.quit:
			if s.l != nil {
				s.l.Close()
			}
			if s.api != nil {
				s.api.Close()
			}
			log.Info("healthCheck quit")
			return
		case <-ticker.C:
			health, err := s.getHealth(sync)
			if err != nil {
				continue
			}
			//sync
			if health {
				if !sync {
					err = s.listen(true)
					if err != nil {
						log.Error("healthCheck ","listen open err",err.Error())
						continue
					}
					sync = true
				}
				unSyncTimes = 0

			} else {
				if sync {
					if unSyncTimes >= unSyncMaxTimes {
						err = s.listen(false)
						if err != nil {
							log.Error("healthCheck ","listen close err",err.Error())
							continue
						}
						sync = false
					}
					unSyncTimes++
				}
			}
		}
	}
}
