//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package observer

import "github.com/docker/docker/client"

// Observer basic interface
type Observer interface {
	Run()
	Stop()

	// Docker client
	Docker() *client.Client
}
