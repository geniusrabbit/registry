//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package consul

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"fmt"

	"github.com/geniusrabbit/registry/service"
	"github.com/hashicorp/consul/api"
)

type discovery struct {
	agent      *api.Agent
	datacenter string
}

// Register new service
func (d *discovery) Register(options service.Options) error {
	var (
		host = options.Address
		port int
	)

	if strings.Contains(host, "://") {
		url, err := url.Parse(host)
		if err != nil {
			return err
		}
		host = url.Host
	}

	if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(host)
		if err != nil {
			return err
		}

		if "" == p {
			return fmt.Errorf("Skip service without port [%s]", options.Address)
		}

		v, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return err
		}

		host = h
		port = int(v)
	}

	return d.agent.ServiceRegister(&api.AgentServiceRegistration{
		ID:                options.ID,
		Name:              options.Name,
		Address:           host,
		Port:              port,
		Tags:              append(options.Tags, "DC="+d.datacenter),
		EnableTagOverride: true,
		Check: &api.AgentServiceCheck{
			Interval: options.Check.Interval,
			Timeout:  options.Check.Timeout,
			HTTP:     options.Check.HTTP,
			TCP:      options.Check.TCP,
			DeregisterCriticalServiceAfter: d.deregisterTime(),
		},
	})
}

// Unregister servece by ID
func (d *discovery) Unregister(id string) error {
	return d.agent.ServiceDeregister(id)
}

// Lookup services by filter
func (d *discovery) Lookup(filter *service.Filter) ([]*service.Service, error) {
	if nil == filter {
		filter = &service.Filter{Datacenter: d.datacenter}
	}

	var statuses = make(map[string]int8)
	if checks, err := d.agent.Checks(); err == nil {
		for _, check := range checks {
			var status int8 = service.StatusUndefined
			switch check.Status {
			case "pass", api.HealthPassing:
				status = service.StatusPassing
			case "warn", api.HealthWarning:
				status = service.StatusWarning
			case "fall", api.HealthCritical:
				status = service.StatusCritical
			}
			statuses[check.ServiceID] = status
		}
	}

	var (
		result        = make([]*service.Service, 0, len(statuses))
		services, err = d.agent.Services()
	)

	if err != nil {
		return nil, err
	}

	for _, srv := range services {
		if srv.Service == "consul" {
			continue
		}

		var (
			status, _ = statuses[srv.ID]
			srv       = &service.Service{
				ID:         srv.ID,
				Name:       srv.Service,
				Datacenter: dc(srv.Tags),
				Address:    srv.Address,
				Port:       srv.Port,
				Tags:       srv.Tags,
				Status:     status,
			}
		)

		if srv.Test(filter) {
			result = append(result, srv)
		}
	}
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
/// Internal methods
///////////////////////////////////////////////////////////////////////////////

func (d *discovery) deregisterTime() string {
	return "10m"
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
