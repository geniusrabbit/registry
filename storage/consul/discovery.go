//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package consul

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/geniusrabbit/registry/service"
	"github.com/hashicorp/consul/api"
)

type discovery struct {
	datacenter string
	agent      *api.Agent
	catalog    *api.Catalog
	health     *api.Health
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
func (d *discovery) Lookup(filter *service.Filter) (services []*service.Service, err error) {
	var dcl []string
	if nil == filter || "*" != filter.Datacenter {
		if nil == filter {
			filter = &service.Filter{Datacenter: d.datacenter}
		}
		return d.lookup(filter)
	}

	if dcl, err = d.catalog.Datacenters(); err != nil {
		return nil, fmt.Errorf("Datacenters list: %s", err)
	}

	for _, dc := range dcl {
		filter.Datacenter = dc
		s, err := d.lookup(filter)
		if err != nil {
			return nil, fmt.Errorf("Datacenter %s lookup: %s", dc, err)
		}
		services = append(services, s...)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////
/// Internal methods
///////////////////////////////////////////////////////////////////////////////

// Lookup services by filter
func (d *discovery) lookup(filter *service.Filter) (result []*service.Service, err error) {
	var (
		names    []string
		list     map[string][]string
		services []*service.Service
		q        = &api.QueryOptions{Datacenter: filter.Datacenter}
	)

	if list, _, err = d.catalog.Services(q); err != nil {
		return nil, err
	}

	for name := range list {
		names = append(names, name)
		items, _, err := d.catalog.Service(name, "", q)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			services = append(services, &service.Service{
				ID:         item.ServiceID,
				Name:       item.ServiceName,
				Datacenter: dc(item.ServiceTags),
				Address:    item.ServiceAddress,
				Port:       item.ServicePort,
				Tags:       item.ServiceTags,
				Status:     service.StatusUndefined,
			})
		}
	}

	for _, name := range names {
		healthChecks, _, err := d.health.Checks(name, q)
		if err != nil {
			return nil, err
		}

		for _, check := range healthChecks {
			f := func(i int) bool { return services[i].ID >= check.ServiceID }
			if i := sort.Search(len(services), f); i > 0 && i < len(services) && services[i].ID == check.ServiceID {
				services[i].Status = status(check.Status)
			}
		}
	}

	for _, srv := range services {
		if srv.Test(filter) {
			result = append(result, srv)
		}
	}
	return
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

func status(st string) (status int8) {
	status = int8(service.StatusUndefined)
	switch st {
	case "pass", api.HealthPassing:
		status = service.StatusPassing
	case "warn", api.HealthWarning:
		status = service.StatusWarning
	case "fall", api.HealthCritical:
		status = service.StatusCritical
	}
	return
}
