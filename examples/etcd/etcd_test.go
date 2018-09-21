// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	endpoints      = []string{"172.16.10.47:2379"}
)

func Client() *clientv3.Client {

	tlsInfo := transport.TLSInfo{
		CertFile:      "./config/certs/etcd.pem",
		KeyFile:       "./config/certs/etcd-key.pem",
		TrustedCAFile: "./config/certs/ca.pem",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}
	return cli

}

func TestEtcdWatch(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	cli := Client()
	var watchList []*clientv3.Event

	//watch key /test/v1
	go func() {
		rch := cli.Watch(context.Background(), "/test/v1")
		for wresp := range rch {
			for _, ev := range wresp.Events {
				watchList = append(watchList, ev)
				log.Debugf("WATCH %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}()

	//put key value
	number := 5
	for i := 0; i < number; i++ {
		_, err := cli.Put(context.Background(), "/test/v1", fmt.Sprintf("test-%d", i))
		if err != nil {
			log.Debugf("Error %v", err)
			return
		}
	}

	//get key /test/v1
	resp, err := cli.Get(context.Background(), "/test/v1")
	if err != nil {
		log.Debugf("Error %v", err)
		return
	}
	for _, k := range resp.Kvs {
		log.Debugf("resp %v", k.Key)
	}

	assert.Equal(t, number, len(watchList))
}
