//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package transport

import (
	"bytes"
	"io/ioutil"
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
	if nil == t.Balancer || '!' != req.URL.Host[0] {
		return t.Transport.RoundTrip(req)
	}

	var body []byte
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
	}

	if conn := t.Balancer.Borrow(req.URL.Host[1:]); nil != conn {
		if req.URL.Host = conn.Host(); len(body) > 0 {
			req.Body = ioutil.NopCloser(
				bytes.NewBuffer(body),
			)
		}
		defer conn.Return(nil)
	}

	return t.Transport.RoundTrip(req)
}
