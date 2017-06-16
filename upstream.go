//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import "math/rand"

// Connect interface
type Connect interface {
	Host() string
	Return(resultError error)
}

type upstreamItem interface {
	Weight() int
	SetWeight(weight int)
	Connect(up *Upstream) Connect
}

// Upstream connection queue
type Upstream struct {
	items       []upstreamItem
	totalWeight int
	stepSize    int
	currentStep int
	queue       chan Connect
}

// NewUpstream queue
func NewUpstream(idleCount int) *Upstream {
	if idleCount < 1 {
		idleCount = 1000
	}
	return &Upstream{
		queue: make(chan Connect, idleCount),
	}
}

// Reset all active streams
func (up *Upstream) Reset() {
	up.stepSize = 0
	up.totalWeight = 0
	for _, it := range up.items {
		it.SetWeight(0)
	}
}

// Update upstream items
func (up *Upstream) Update(items ...upstreamItem) {
	for _, item := range items {
		if idx, it := up.itemByHost(item.Connect(up).Host()); nil != it {
			up.items[idx] = item
		} else {
			up.items = append(up.items, item)
		}
	}

	// Recalck step item
	up.refreshStepCounters()
}

// Borrow next connection
func (up *Upstream) Borrow() Connect {
	select {
	case conn := <-up.queue:
		return conn
	default:
		return up.Next()
	}
}

// Return connection into queue
func (up *Upstream) Return(conn Connect, resultError error) {
	if nil != conn && nil == resultError {
		select {
		case up.queue <- conn:
		default:
		}
	}
}

// Next connection
func (up *Upstream) Next() Connect {
	if len(up.items) > 0 {
		if up.stepSize > 0 && up.totalWeight > 0 {
			var row = up.nextStep()
			for _, it := range up.items {
				if it.Weight() > row {
					return it.Connect(up)
				}
				row -= it.Weight()
			}
		}
		return up.items[rand.Intn(len(up.items)-1)].Connect(up)
	}
	return nil
}

// itemByHost for upstgream
func (up *Upstream) itemByHost(host string) (int, upstreamItem) {
	for i, it := range up.items {
		if it.Connect(up).Host() == host {
			return i, it
		}
	}
	return -1, nil
}

func (up *Upstream) nextStep() int {
	if up.totalWeight > 0 {
		up.currentStep = (up.currentStep + up.stepSize) % up.totalWeight
	} else {
		up.currentStep = 0
	}
	return up.currentStep
}

func (up *Upstream) refreshStepCounters() {
	up.totalWeight, up.stepSize = 0, 0
	for _, item := range up.items {
		up.totalWeight += item.Weight()
		if up.stepSize < item.Weight() {
			up.stepSize = item.Weight()
		}
	}

	if up.stepSize > 0 && up.totalWeight%up.stepSize == 0 {
		up.stepSize++
	}
}
