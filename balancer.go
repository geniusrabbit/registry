//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import (
	"sync"
	"time"

	"github.com/geniusrabbit/registry/service"
)

// Balancer of service
type Balancer interface {
	// Borrow service from upstream
	Borrow(service string) Connect

	// Return connect back to pool
	Return(conn Connect, errResult error)
}

type balancer struct {
	sync.Mutex
	ticker            *time.Ticker
	maxIdelConnection int
	discovery         service.Discovery
	serviceStreams    map[string]*Upstream
}

// NewBalancer object
func NewBalancer(discovery service.Discovery, maxIdelConnection int) Balancer {
	if nil == discovery {
		panic("Undefined discovery service")
	}
	return &balancer{
		maxIdelConnection: maxIdelConnection,
		discovery:         discovery,
		serviceStreams:    map[string]*Upstream{},
	}
}

// Borrow service from upstream
func (b *balancer) Borrow(service string) Connect {
	if upst, ok := b.serviceStreams[service]; ok {
		return upst.Borrow()
	}
	return nil
}

// Return connect back to pool
func (b *balancer) Return(conn Connect, errResult error) {
	conn.Return(errResult)
}

// ServiceError event
func (b *balancer) Supervisord(interval time.Duration) {
	b.Lock()

	if nil != b.ticker {
		b.ticker.Stop()
		b.ticker = time.NewTicker(interval)
	}

	b.Unlock()

	for {
		select {
		case <-b.ticker.C:
			b.refresh()
		default:
			return
		}
	}
}

// Stop supervisord
func (b *balancer) Stop() {
	b.Lock()
	if nil != b.ticker {
		b.ticker.Stop()
		b.ticker = nil
	}
	b.Unlock()
}

///////////////////////////////////////////////////////////////////////////////
/// Internal methods
///////////////////////////////////////////////////////////////////////////////

func (b *balancer) refresh() error {
	services, err := b.discovery.Lookup(nil)
	if len(services) < 1 || nil != err {
		return err
	}

	for _, up := range b.serviceStreams {
		up.Reset()
	}

	for _, srv := range services {
		if upstream, ok := b.serviceStreams[srv.Name]; ok {
			upstream.Update(UpstreamService(srv))
		} else {
			upstream = NewUpstream(b.maxIdelConnection)
			upstream.Update(UpstreamService(srv))
			b.serviceStreams[srv.Name] = upstream
		}
	}
	return nil
}
