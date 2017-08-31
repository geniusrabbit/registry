//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package transport

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/geniusrabbit/registry"
)

// Transport wrapper for HTTP connection to service
type Transport struct {
	http.Transport
	registry.Balancer
}

// RoundTrip of HTTP request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Balancer == nil || '!' != req.URL.Host[0] {
		return t.Transport.RoundTrip(req)
	}

	if req.Body != nil {
		if body, _ := ioutil.ReadAll(req.Body); len(body) > 0 {
			req.Body = ioutil.NopCloser(
				bytes.NewBuffer(body),
			)
		}
	}

	req.URL.Host = req.URL.Host[1:]
	host, _, _ := net.SplitHostPort(req.URL.Host)
	if len(host) < 1 {
		host = req.URL.Host
	}

	if conn := t.Balancer.Borrow(host); conn != nil {
		req.URL.Host = conn.Host()
		defer conn.Return(nil)
	}

	return t.Transport.RoundTrip(req)
}
