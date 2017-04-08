//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package docker

// Stats of container usage
type Stats struct {
	CPUUsage    float64 // 0-100
	MemoryUsage uint64  // bytes
	MemoryLimit uint64  // bytes
}
