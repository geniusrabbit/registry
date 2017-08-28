//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

// +build ignore

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/demdxx/gocast"
	"github.com/docker/docker/client"
	"github.com/geniusrabbit/registry/observer"
	"github.com/geniusrabbit/registry/observer/docker"
	"github.com/geniusrabbit/registry/service"
	"github.com/geniusrabbit/registry/storage/consul"
)

var (
	flagRegistry = flag.String("r", "", "Consul connect URL (default env REGISTRY_DNS)")
)

func init() {
	flag.Parse()

	var formatter log.Formatter

	if log.IsTerminal() {
		formatter = &log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05 MST",
		}
	} else {
		formatter = &log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05 MST",
		}
	}

	log.SetLevel(log.InfoLevel)
	if gocast.ToBool(os.Getenv("DEBUG")) {
		log.SetLevel(log.DebugLevel)
		go func() { log.Println(http.ListenAndServe(":6060", nil)) }()
	}

	log.SetFormatter(formatter)
}

func main() {
	var consulRegistry = *flagRegistry
	if consulRegistry == "" {
		consulRegistry = os.Getenv("REGISTRY_DNS")
	}

	fmt.Println("> Connect to:", consulRegistry)
	if storage, err := consul.New("", consulRegistry); nil == err {
		newObserver(storage.Discovery()).Run()
	} else {
		log.Error(err)
	}
}

// Observer event processor
type obs struct {
	docker    *client.Client
	discovery service.Discovery
}

func newObserver(discovery service.Discovery) observer.Observer {
	var (
		subObs   = &obs{discovery: discovery}
		obs, err = docker.New(
			subObs,
			os.Getenv("DOCKER_HOST"),
			os.Getenv("DOCKER_API_VERSION"),
			nil, nil,
		)
	)
	if nil != err {
		panic(err)
	}
	subObs.docker = obs.Docker()
	return obs
}

func (o *obs) Event(containerID, action string) {
	switch action {
	case "start", "unpause", "refresh":
		log.Debugf("Register service: %s", containerID[:12])
		go func() {
			if err := o.serviceRegister(containerID); err != nil {
				log.Errorf("Register service [%s]: %v", containerID[:12], err)
			}
		}()
	case "die", "kill", "stop", "pause", "oom":
		log.Debugf("Unregister service [%s]: %s", action, containerID[:12])
		if err := o.discovery.Unregister(containerID); err != nil {
			log.Errorf("Deregister service [%s]: %v", action, err)
		}
	}
}

func (o *obs) Error(err error) {
	log.Errorf("Event: %v", err)
}

func (o *obs) serviceRegister(containerID string) error {
	service, err := docker.ServiceInfo(containerID, o.docker)
	if nil != service && nil == err {
		err = o.discovery.Register(*service)
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func def(v, def string) string {
	if len(v) < 1 {
		return def
	}
	return v
}
