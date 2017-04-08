//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package service

import (
	"net"
	"strconv"
	"strings"
)

// Options of service
type Options struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Tags    []string  `json:"tags,omitempty"`
	Check   CheckInfo `json:"check,omitempty"`
}

// CheckInfo of service state
type CheckInfo struct {
	Interval string `json:"interval,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	HTTP     string `json:"http,omitempty"`
	TCP      string `json:"tcp,omitempty"`
}

// AddTag to options
func (o *Options) AddTag(key, value string) {
	o.Tags = append(o.Tags, key+"="+value)
}

// Service from options
func (o *Options) Service() *Service {
	var (
		host, port, _ = net.SplitHostPort(o.Address)
		portInt, _    = strconv.ParseInt(port, 10, 64)
	)

	return &Service{
		ID:         o.ID,
		Name:       o.Name,
		Datacenter: dc(o.Tags),
		Address:    host,
		Port:       int(portInt),
		Tags:       o.Tags,
		Status:     StatusUndefined,
		weight:     WeightByTags(o.Tags),
	}
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func dc(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "DC=") {
			return strings.TrimPrefix(tag, "DC=")
		}
	}
	return ""
}
