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
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	flagListen   = flag.String("listen", ":8080", "Listen and serve HTTP address")
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
		go runWebService(*flagListen, storage)
		newObserver(storage.Discovery()).Run()
	} else {
		log.Error(err)
	}
}

func runWebService(address string, storage *consul.Storage) error {
	srv := echo.New()

	srv.Use(middleware.Logger())
	srv.Use(middleware.Recover())
	srv.Use(middleware.CORS())

	srv.GET("/v1/unregister/:service", unregisterService(storage))
	srv.GET("/healthcheck", healthCheck)

	return srv.Start(address)
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
	case "die", "kill", "stop", "pause", "oom", "destroy":
		log.Debugf("Unregister service [%s]: %s", action, containerID[:12])
		if err := o.discovery.Unregister(containerID); err != nil {
			log.Errorf("Deregister service [%s]: %v", action, err)
		}

		// Unregister swarm service
		if err := o.discovery.Unregister("service:" + containerID); err != nil {
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

func unregisterService(storage *consul.Storage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			discovery  = storage.Discovery()
			servs, err = discovery.Lookup(&service.Filter{Service: ctx.Param("service")})
		)

		if err != nil {
			return ctx.JSON(http.StatusOK, map[string]interface{}{
				"result": "error",
				"error":  err,
			})
		}

		for _, srv := range servs {
			discovery.Unregister(srv.ID)
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"result": "ok",
		})
	}
}

func healthCheck(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
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
