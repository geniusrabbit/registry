//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
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
	eventer  ServiceContainerEventer
	observer *baseObserver
}

// NewService for current docker container
func NewService(eventer ServiceContainerEventer, host, version string, httpClient *http.Client, httpHeader map[string]string) (observer.Observer, error) {
	var (
		self     = &serviceObserver{}
		obs, err = New(self, host, version, httpClient, httpHeader)
	)
	if nil == err {
		if obs, _ := obs.(*baseObserver); nil != obs {
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
	if options, err := ServiceInfo(containerID, s.observer.docker); nil == err {
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
	if nil != err {
		s.eventer.ServiceError(err)
	}
}
