//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package service

// Discovery service
type Discovery interface {
	// Register new service
	Register(options Options) error

	// Unregister servece by ID
	Unregister(id string) error

	// Lookup services by filter
	Lookup(filter *Filter) ([]*Service, error)
}
