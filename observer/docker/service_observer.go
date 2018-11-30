//
// @project registry 2017 - 2018
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017 - 2018
//

package docker

import (
	"net/http"

	"github.com/docker/docker/client"
	"github.com/geniusrabbit/registry/observer"
	"github.com/geniusrabbit/registry/service"
)

// ServiceContainerEventer processor
type ServiceContainerEventer interface {
	ServiceEvent(event string, srv *service.Service)
	ServiceError(err error)
}

type serviceObserver struct {
	hostIPAddr bool
	eventer    ServiceContainerEventer
	observer   *baseObserver
}

// NewService for current docker container
func NewService(eventer ServiceContainerEventer, host, version string, httpClient *http.Client, httpHeader map[string]string, registerHost bool) (observer.Observer, error) {
	var (
		self     = &serviceObserver{eventer: eventer, hostIPAddr: registerHost}
		obs, err = New(self, host, version, httpClient, httpHeader)
	)
	if err == nil {
		if obs, _ := obs.(*baseObserver); obs != nil {
			self.observer = obs
			return self, nil
		}
	}
	return nil, err
}

// Run observer
func (s *serviceObserver) Run() {
	s.observer.Run()
}

// Stop observer
func (s *serviceObserver) Stop() {
	s.observer.Stop()
}

// Docker client
func (s *serviceObserver) Docker() *client.Client {
	return s.observer.Docker()
}

// Event processor
func (s *serviceObserver) Event(containerID, action string) {
	options, err := ServiceInfo(containerID, s.hostIPAddr, s.observer.docker)
	if err == nil {
		var srv = options.Service()
		switch action {
		case "start", "unpause", "refresh":
			srv.Status = service.StatusPassing
		case "stop", "pause":
			srv.Status = service.StatusWarning
		case "die", "kill", "oom":
			srv.Status = service.StatusCritical
		}
		s.eventer.ServiceEvent(action, srv)
	} else {
		s.Error(err)
	}
}

// Error processor
func (s *serviceObserver) Error(err error) {
	if err != nil {
		s.eventer.ServiceError(err)
	}
}
