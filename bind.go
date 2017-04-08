//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

type bind struct {
	storages []Storage
	target   bindUpdater
}

func (b *bind) subscribe(storages []Storage) {
	for _, s := range storages {
		var sub = true

		for _, ss := range b.storages {
			if ss == s {
				sub = false
				break
			}
		}

		if sub {
			s.Subscribe(func(key string, val interface{}) {
				if nil == b {
					return
				}
				b.target.UpdateKey(key, val)
			})
			b.storages = append(b.storages, s)
		} // end if
	}
}
