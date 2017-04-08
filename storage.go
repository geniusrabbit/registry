//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import (
	"time"

	"github.com/geniusrabbit/registry/service"
)

// Storage connector
type Storage interface {
	// Subscribe config key updater
	Subscribe(f func(key string, value interface{}))

	// Discovery services
	Discovery() service.Discovery

	// Supervisor of auto refresh
	Supervisor(interval time.Duration)
}
