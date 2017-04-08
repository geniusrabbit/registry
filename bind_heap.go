//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import "time"

// BindHeap base type
type BindHeap struct {
	storages []Storage
	binds    []bind
}

// RegisterStore in heap
func (bh *BindHeap) RegisterStore(st Storage) {
	bh.storages = append(bh.storages, st)
}

// Bind config for authoupdate
func (bh *BindHeap) Bind(conf interface{}, sep ...string) error {
	var bn bind

	if i, ok := conf.(bindUpdater); ok {
		bn = bind{target: i}
	} else {
		var (
			tagName = "store"
			pathSep = rune('/')
		)

		if len(sep) > 0 && len(sep[0]) > 0 {
			tagName = sep[0]
		}

		if len(sep) > 1 && len(sep[1]) > 0 {
			pathSep = rune(sep[1][0])
		}

		bn = bind{
			target: WrapConf(conf, tagName, pathSep),
		}
	}

	bh.binds = append(bh.binds, bn)
	bn.subscribe(bh.storages)

	return nil
}

// Subscribe updater for config
func (bh *BindHeap) Subscribe(conf interface{}, key string, b bindKeyUpdater) error {
	if up := bh.confUpdater(conf); nil != up {
		up.Subscribe(key, b)
		return nil
	}
	return ErrUnbindedConfig
}

// Unsubscribe updater for config
func (bh *BindHeap) Unsubscribe(conf interface{}, key string, b bindKeyUpdater) error {
	if up := bh.confUpdater(conf); nil != up {
		up.Unsubscribe(key, b)
		return nil
	}
	return ErrUnbindedConfig
}

// Supervisor of auto refresh
func (bh *BindHeap) Supervisor(interval time.Duration) {
	for _, st := range bh.storages {
		go st.Supervisor(interval)
	}
}

func (bh *BindHeap) confUpdater(conf interface{}) bindUpdater {
	for _, bn := range bh.binds {
		if bn.target == conf {
			return bn.target
		}

		if bw, _ := bn.target.(*bindWrapper); nil != bw {
			if bw.target == conf {
				return bw
			}
		}
	}
	return nil
}

// RegisterStore in heap
func RegisterStore(st Storage) {
	DefaultBindHeap.RegisterStore(st)
}

// Bind config for authoupdate
func Bind(conf interface{}, sep ...string) error {
	return DefaultBindHeap.Bind(conf, sep...)
}

// Subscribe updater for config
func Subscribe(conf interface{}, key string, b bindKeyUpdater) error {
	return DefaultBindHeap.Subscribe(conf, key, b)
}

// Unsubscribe updater for config
func Unsubscribe(conf interface{}, key string, b bindKeyUpdater) error {
	return DefaultBindHeap.Unsubscribe(conf, key, b)
}

// Supervisor of auto refresh
func Supervisor(interval time.Duration) {
	DefaultBindHeap.Supervisor(interval)
}

// Global vars
var (
	DefaultBindHeap BindHeap
)
