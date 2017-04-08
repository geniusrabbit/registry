//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry_test

import (
	"testing"

	"io"

	registry "."
	"github.com/stretchr/testify/assert"
)

type conf struct {
	Service string `store:"service/name"`
	IP      string `store:"service/ip"`
}

type store struct {
	fn func(key string, value interface{})
}

func (s *store) Subscribe(f func(key string, value interface{})) {
	s.fn = f
	s.fn("service/name", "test")
	s.fn("service/ip", "127.0.0.1")
}

type keyUpdater struct{}

func (ku *keyUpdater) ConfigKeyUpdate(target interface{}, key string, value interface{}) error {
	if key == "service/ip" {
		target.(*conf).IP = "192.168.0.1"
		return io.EOF
	}
	return nil
}

func TestBind(t *testing.T) {
	var (
		st         = &store{}
		conf       = &conf{}
		keyUpdater = &keyUpdater{}
	)
	registry.RegisterStore(st)
	registry.Bind(conf, "store", "/")

	assert.True(t, "test" == conf.Service, "Invalid service name")
	assert.True(t, "127.0.0.1" == conf.IP, "Invalid ip address")

	// Test change field
	st.fn("service/ip", "127.0.0.2")
	assert.True(t, "127.0.0.2" == conf.IP, "Invalid ip2 address")

	// Test key subscriber
	assert.NoError(t, registry.Subscribe(conf, "service", keyUpdater), "Subscribe")

	st.fn("service/ip", "127.0.0.1")
	assert.True(t, "192.168.0.1" == conf.IP, "Invalid ip3 address")

	// Test key unsubscribe
	assert.NoError(t, registry.Unsubscribe(conf, "service", keyUpdater), "Unsubscribe")

	st.fn("service/ip", "127.0.0.1")
	assert.True(t, "127.0.0.1" == conf.IP, "Invalid ip4 address")
}
