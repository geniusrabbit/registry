//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package transport

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/geniusrabbit/registry/service"
)

// Transport wrapper for HTTP connection to service
type Transport struct {
	http.Transport
	service.Balancer
}

// RoundTrip of HTTP request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if nil == t.Balancer || '!' != req.URL.Host[0] {
		return t.Transport.RoundTrip(req)
	}

	var body []byte
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
	}

	if srv, err := t.Balancer.Borrow(req.URL.Host[1:]); nil == err {
		if req.URL.Host = srv.Host(); len(body) > 0 {
			req.Body = ioutil.NopCloser(
				bytes.NewBuffer(body),
			)
		}
		defer t.Balancer.Release(srv)
	}

	return t.Transport.RoundTrip(req)
}
