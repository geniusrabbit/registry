//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package service

// Filter for search service
type Filter struct {
	ID         string
	Status     int8
	Tags       []string
	Service    string
	Datacenter string
}
