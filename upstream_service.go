//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import "github.com/geniusrabbit/registry/service"

// UpstreamServiceItem wrapper
type UpstreamServiceItem struct {
	host     string
	Service  *service.Service
	Upstream *Upstream
}

// UpstreamService wrapper function
func UpstreamService(srv *service.Service) *UpstreamServiceItem {
	return &UpstreamServiceItem{Service: srv}
}

// Connect service interface
func (it *UpstreamServiceItem) Connect(up *Upstream) Connect {
	it.Upstream = up
	return it
}

// Host name with port
func (it *UpstreamServiceItem) Host() string {
	if len(it.host) < 1 {
		it.host = it.Service.Host()
	}
	return it.host
}

// Weight of service
func (it *UpstreamServiceItem) Weight() int {
	return it.Service.Weight()
}

// SetWeight of service
func (it *UpstreamServiceItem) SetWeight(weight int) {
	it.Service.SetWeight(weight)
}

// Return service to upstream pool
func (it *UpstreamServiceItem) Return(resultError error) {
	if it.Service.Weight() > 0 {
		it.Upstream.Return(it, resultError)
	}
}
