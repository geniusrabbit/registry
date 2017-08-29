//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/geniusrabbit/registry/service"
)

// ServiceInfo by container ID
func ServiceInfo(containerID string, docker *client.Client) (*service.Options, error) {
	var (
		container, err = docker.ContainerInspect(
			context.Background(),
			containerID,
		)
		ipAddr = container.NetworkSettings.IPAddress
	)

	if err == nil && len(ipAddr) < 1 {
		ipAddr, err = resolveLocalIP()
	}

	if err != nil {
		return nil, err
	}

	if container.State.Status != "running" {
		return nil, fmt.Errorf("Container [%s] is not running", containerID[:12])
	}

	var (
		port string
		name string
		tags []string
	)

	// Get tags from environment
	for _, env := range container.Config.Env {
		switch {
		case strings.HasPrefix(env, "SERVICE_NAME="):
			name = strings.TrimPrefix(env, "SERVICE_NAME=")
		case strings.HasPrefix(env, "SERVICE_PORT="):
			port = strings.TrimPrefix(env, "SERVICE_PORT=")
		case strings.HasPrefix(env, "TAG_"):
			tags = append(tags, strings.TrimPrefix(env, "TAG_"))
		}
	}

	// Get tags from labels
	for label, val := range container.Config.Labels {
		switch {
		case label == "service.name":
			name = val
		case label == "service.port":
			port = val
		case strings.HasPrefix(label, "service.tag_"):
			tags = append(tags, strings.TrimPrefix(label, "TAG_")+"="+val)
		}
	}

	if len(name) < 1 {
		return nil, fmt.Errorf("Container [%s] is not the service", containerID[:12])
	}

	// Get port if not defined
	if len(port) < 1 && len(container.NetworkSettings.Ports) > 0 {
		for keyPort, targetPort := range container.NetworkSettings.Ports {
			if len(targetPort) > 0 {
				port = targetPort[0].HostPort
				break
			} else if len(keyPort) > 0 {
				if pl := strings.Split(string(keyPort), "/"); len(pl) > 0 {
					port = pl[0]
					break
				}
			}
		}
	}

	// Container stat
	stats, err := ContainerStats(containerID, docker)
	if err != nil {
		return nil, err
	}

	tags = append(tags,
		fmt.Sprintf("CPU_USAGE=%f", stats.CPUUsage),
		fmt.Sprintf("MEMORY_USAGE=%f", (float64(stats.MemoryUsage)/float64(stats.MemoryLimit))*100),
		fmt.Sprintf("MEMORY_LIMIT=%d", stats.MemoryLimit),
		fmt.Sprintf("PORT_MAP=%s", toJSON(container.NetworkSettings.Ports)),
	)

	return &service.Options{
		ID:      container.ID,
		Name:    name,
		Tags:    tags,
		Address: ipAddr + ":" + port,
		Check: checkOptions(
			ipAddr+":"+port,
			container.Config.Env,
		),
	}, nil
}

// ContainerStats information
func ContainerStats(containerID string, docker *client.Client) (*Stats, error) {
	var (
		stats         types.Stats
		response, err = docker.ContainerStats(context.Background(), containerID, false)
	)

	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(response.Body).Decode(&stats); err != nil {
		return nil, err
	}

	var (
		cpuUsage    float64
		cpuDelta    = float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta = float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuUsage = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	return &Stats{
		CPUUsage:    cpuUsage,
		MemoryUsage: stats.MemoryStats.Usage,
		MemoryLimit: stats.MemoryStats.Limit,
	}, nil
}

func checkOptions(address string, env []string) service.CheckInfo {
	options := service.CheckInfo{
		Interval: "5s",
		Timeout:  "2s",
	}

	for _, e := range env {
		switch {
		case strings.HasPrefix(e, "CHECK_INTERVAL="):
			options.Interval = strings.TrimPrefix(e, "CHECK_INTERVAL=")
		case strings.HasPrefix(e, "CHECK_TIMEOUT="):
			options.Timeout = strings.TrimPrefix(e, "CHECK_TIMEOUT=")
		case strings.HasPrefix(e, "CHECK_HTTP="):
			options.HTTP = strings.Replace(strings.TrimPrefix(e, "CHECK_HTTP="), "{{address}}", address, 1)
		case strings.HasPrefix(e, "CHECK_TCP="):
			options.TCP = strings.Replace(strings.TrimPrefix(e, "CHECK_TCP="), "{{address}}", address, 1)
		}
	}
	return options
}

func toJSON(v interface{}) string {
	if json, err := json.Marshal(v); err == nil {
		return string(json)
	}
	return ""
}

// resolveLocalIP returns the non loopback local IP of the host
func resolveLocalIP() (string, error) {
	if hostIP := os.Getenv("DOCKER_HOST_IP"); len(hostIP) > 0 {
		return hostIP, nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
