//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Service statuses
const (
	StatusUndefined = iota
	StatusPassing
	StatusWarning
	StatusCritical
)

// Service info
type Service struct {
	ID         string
	Name       string
	Datacenter string
	Address    string
	Port       int
	Tags       []string
	Status     int8
	weight     int
}

// Host including port
func (s *Service) Host() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

// Weight of service
func (s *Service) Weight() int {
	if s.Status != StatusPassing {
		return 0
	}
	if s.weight < 1 {
		s.weight = 1
	}
	return s.weight
}

// SetWeight of service
func (s *Service) SetWeight(weight int) {
	if weight < 1 {
		s.Status = StatusUndefined
	}
	s.weight = weight
}

// Test service in comparison with filter
func (s *Service) Test(filter *Filter) bool {
	if filter == nil {
		return true
	}

	if len(filter.ID) != 0 && filter.ID != s.ID {
		return false
	}

	if filter.Status > 0 && filter.Status != s.Status {
		return false
	}

	if len(filter.Datacenter) != 0 && filter.Datacenter != "*" && filter.Datacenter != s.Datacenter {
		return false
	}

	if len(filter.Service) != 0 && filter.Service != s.Name {
		return false
	}

	if len(filter.Tags) != 0 {
		for _, ft := range filter.Tags {
			for _, st := range s.Tags {
				if ft == st {
					return true
				}
			}
		}
		return false
	}
	return true
}

// WeightByTags by tags
func WeightByTags(tags []string) int {
	var (
		cpuUsage float64
		memUsage float64
	)

	for _, tag := range tags {
		switch {
		case strings.HasPrefix("CPU_USAGE=", tag):
			cpuUsage, _ = strconv.ParseFloat(tag[10:], 64)
		case strings.HasPrefix("MEMORY_USAGE=", tag):
			memUsage, _ = strconv.ParseFloat(tag[13:], 64)
		}
	}

	return int(1000.0 * ((memUsage/2.0 + cpuUsage) / 150.0))
}

// List imprimentation of sort.Sorter interface
type List []*Service

func (a List) Len() int           { return len(a) }
func (a List) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a List) Less(i, j int) bool { return a[i].ID < a[j].ID }
func (a List) Sort()              { sort.Sort(a) }
