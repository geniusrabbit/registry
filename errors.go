//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import "errors"

// Errors set
var (
	ErrUnbindedConfig      = errors.New("Config is not bind")
	ErrInvalidKeyParam     = errors.New("Invalid key params")
	ErrInvalidTargetStruct = errors.New("Invalid bind target struct")
)
