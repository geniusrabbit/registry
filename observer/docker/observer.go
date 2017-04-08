//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package docker

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/geniusrabbit/registry/observer"
	"github.com/labstack/gommon/log"
)

type ContainerEventer interface {
	Event(containerID, action string)
	Error(err error)
}

// Service container observer
type baseObserver struct {
	sync.Mutex
	ContainerEventer
	inProcess bool
	ticker    *time.Ticker
	docker    *client.Client
}

// New for current docker container
func New(eventer ContainerEventer, host, version string, httpClient *http.Client, httpHeader map[string]string) (observer.Observer, error) {
	client, err := client.NewClient(
		def(host, client.DefaultDockerHost),
		def(version, api.DefaultVersion),
		httpClient,
		httpHeader,
	)

	if nil != err {
		return nil, err
	}

	return &baseObserver{
		ContainerEventer: eventer,
		docker:           client,
	}, nil
}

// Run baseObserver
func (o *baseObserver) Run() {
	o.Stop()

	o.ticker = time.NewTicker(30 * time.Second)
	var messages, errors = o.docker.Events(context.Background(), types.EventsOptions{})

	for {
		select {
		case msg := <-messages:
			if events.ContainerEventType == msg.Type {
				o.ContainerEventer.Event(msg.Actor.ID, msg.Action)
			}
		case err := <-errors:
			if nil != err {
				o.ContainerEventer.Error(err)
			}
		case <-o.ticker.C:
			o.refreshAll()
		}
	}
}

func (o *baseObserver) Stop() {
	if nil != o.ticker {
		o.ticker.Stop()
		o.ticker = nil
	}
}

// Docker client
func (o *baseObserver) Docker() *client.Client {
	return o.docker
}

func (o *baseObserver) refreshAll() {
	if o.goInProcess() {
		return
	}

	containers, err := o.docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Errorf("Refresh services (container list): %v", err)
		return
	}

	for _, container := range containers {
		o.ContainerEventer.Event(container.ID, "refresh")
	}

	o.outOfProcess()
}

func (o *baseObserver) goInProcess() bool {
	o.Lock()
	defer o.Unlock()

	if o.inProcess {
		return true
	}

	o.inProcess = true
	return false
}

func (o *baseObserver) outOfProcess() {
	o.Lock()
	o.inProcess = false
	o.Unlock()
}

func def(v, def string) string {
	if len(v) > 0 {
		return v
	}
	return def
}
