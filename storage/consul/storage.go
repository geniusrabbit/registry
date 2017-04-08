//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package consul

import (
	"net/url"
	"sync"
	"time"

	"github.com/geniusrabbit/registry/service"
	"github.com/hashicorp/consul/api"
)

// Storage accessor to consul
type Storage struct {
	sync.Mutex
	prefix      string
	datacenter  string
	client      *api.Client
	subscribers []func(key string, value interface{})
	ticker      *time.Ticker
}

// New storage connector
func New(prefix, link string) (*Storage, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(&api.Config{
		Scheme:     url.Scheme,
		Address:    url.Host,
		Datacenter: url.Path[1:],
		Token:      url.Query().Get("token"),
	})

	if err != nil {
		return nil, err
	}

	return &Storage{
		prefix:     prefix,
		datacenter: url.Path[1:],
		client:     client,
	}, nil
}

// Subscribe config key updater
func (s *Storage) Subscribe(f func(key string, value interface{})) {
	s.subscribers = append(s.subscribers, f)
}

// Discovery services
func (s *Storage) Discovery() service.Discovery {
	return &discovery{
		agent:      s.client.Agent(),
		datacenter: s.datacenter,
	}
}

// Supervisor of auto refresh
func (s *Storage) Supervisor(interval time.Duration) {
	s.Stop()
	s.ticker = time.NewTicker(interval)

	for {
		select {
		case <-s.ticker.C:
			s.refresh()
		}
	} // end for
}

// Stop supervisord
func (s *Storage) Stop() {
	if nil != s.ticker {
		s.Lock()
		s.ticker.Stop()
		s.ticker = nil
		s.Unlock()
	}
}

// key value accessor
func (s *Storage) kv() kv {
	return kv{client: s.client.KV(), prefix: s.prefix}
}

// refresh subscribed events
func (s *Storage) refresh() {
	if len(s.subscribers) < 0 {
		return
	}

	var data, _ = s.kv().List(s.prefix)
	if nil == data {
		return
	}

	for _, sub := range s.subscribers {
		for key, val := range data {
			sub(key, val)
		}
	}
}
