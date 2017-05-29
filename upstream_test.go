//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	registry "."
)

type item struct {
	Port int
}

func (i *item) Weight() int {
	return i.Port % 21
}

func (i *item) SetWeight(weight int) {
	// DUMMY....
}

func (i *item) Connect(_ *registry.Upstream) registry.Connect {
	return i
}

func (i *item) Return(resultError error) {}

func (i *item) Host() string {
	return fmt.Sprintf("host:%04d => %d", i.Port, i.Weight())
}

func TestUpstream(t *testing.T) {
	var (
		wg sync.WaitGroup
		up = registry.NewUpstream(1000)
	)

	up.Update(
		&item{Port: 1000},
		&item{Port: 700},
		&item{Port: 192},
	)

	wg.Add(10000)

	for i := 0; i < 10000; i++ {
		conn := up.Borrow()
		if nil == conn {
			fmt.Println(i, "Nope!")
		} else {
			fmt.Println(i, conn.Host())
		}

		go func() {
			time.Sleep(time.Millisecond * 300)
			up.Return(conn, nil)
			wg.Done()
		}()
	}

	wg.Wait()
}
