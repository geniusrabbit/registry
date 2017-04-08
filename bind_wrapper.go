//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package registry

import (
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/demdxx/gocast"
)

type bindKeyUpdater interface {
	ConfigKeyUpdate(target interface{}, key string, value interface{}) error
}

type bindUpdater interface {
	UpdateKey(key string, value interface{}) error
	Subscribe(key string, b bindKeyUpdater)
	Unsubscribe(key string, b bindKeyUpdater)
}

type bindWrapper struct {
	sync.Mutex
	previousData map[string]*string
	subscribe    map[string][]bindKeyUpdater
	tagSeparator rune
	tagName      string
	target       interface{}
}

// WrapConf bind field updater
func WrapConf(conf interface{}, tagName string, pathSeparator rune) bindUpdater {
	return &bindWrapper{
		previousData: map[string]*string{},
		tagSeparator: pathSeparator,
		tagName:      tagName,
		target:       conf,
	}
}

// UpdateKey in config
func (wr *bindWrapper) UpdateKey(key string, value interface{}) error {
	if "" == key {
		return ErrInvalidKeyParam
	}

	if nil != wr.subscribe {
		var end = false

		for baseKey, subs := range wr.subscribe {
			if key == baseKey || strings.HasPrefix(key, baseKey) {
				for _, sub := range subs {
					if err := sub.ConfigKeyUpdate(wr.target, key, value); nil != err {
						if io.EOF == err {
							end = true
						} else {
							return err
						}
					}
				} // end for
			}
		}

		if end {
			return nil
		}
	}

	if !wr.upAndContinue(key, value) {
		return nil
	}

	return wr.setStructItem(
		reflect.ValueOf(wr.target),
		value,
		key,
		strings.Split(key, string(wr.tagSeparator))...)
}

// Subscribe key updater
func (wr *bindWrapper) Subscribe(key string, b bindKeyUpdater) {
	if nil == wr.subscribe {
		wr.subscribe = map[string][]bindKeyUpdater{}
	}

	wr.Lock()
	defer wr.Unlock()

	subs, _ := wr.subscribe[key]
	wr.subscribe[key] = append(subs, b)
}

// Unsubscribe key updater
func (wr *bindWrapper) Unsubscribe(key string, b bindKeyUpdater) {
	if nil == wr.subscribe {
		return
	}

	wr.Lock()
	defer wr.Unlock()

	if subs, _ := wr.subscribe[key]; len(subs) > 0 {
		var newSubs []bindKeyUpdater

		for _, s := range subs {
			if s != b {
				newSubs = append(newSubs, s)
			}
		} // end for

		if len(newSubs) != len(subs) {
			wr.subscribe[key] = newSubs
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
/// Intenal methods
///////////////////////////////////////////////////////////////////////////////

func (wr *bindWrapper) fieldTagName(field reflect.StructField) string {
	return field.Tag.Get(wr.tagName)
}

func (wr *bindWrapper) setStructItem(ref reflect.Value, value interface{}, key string, keys ...string) (err error) {
	if ref = unpoint(ref); !ref.IsValid() || (ref.Kind() == reflect.Ptr && ref.IsNil()) {
		return ErrInvalidTargetStruct
	}

	var tp = ref.Type()
	for i := 0; i < tp.NumField(); i++ {
		var (
			field = tp.Field(i)
			tag   = wr.fieldTagName(field)
		)

		if tag == "" {
			err = wr.setStructItem(ref.Field(i), value, key, keys...)
		} else {
			if newKey, set := wr.prepareKey(key, tag, keys); set {
				err = wr.setValue(ref.Field(i), value)
			} else if len(newKey) > 0 {
				err = wr.setStructItem(ref.Field(i), value, key, newKey...)
			}
		}

		if nil != err {
			break
		}
	}
	return
}

func (wr *bindWrapper) prepareKey(baseKey, curKey string, keys []string) ([]string, bool) {
	if wr.tagSeparator == rune(curKey[0]) {
		curKey = curKey[1:]
	} else {
		baseKey = strings.Join(keys, string(wr.tagSeparator))
	}

	if strings.HasPrefix(baseKey, curKey) {
		if curKey = baseKey[len(curKey)-1:]; len(curKey) > 1 {
			if wr.tagSeparator == rune(curKey[0]) {
				return strings.Split(curKey, string(wr.tagSeparator)), false
			}
		} else {
			return nil, true
		}
	}
	return nil, false
}

func (wr *bindWrapper) setValue(ref reflect.Value, value interface{}) (err error) {
	if nil == value {
		ref.Set(reflect.Zero(ref.Type()))
	} else {
		var val interface{}
		if val, err = gocast.ToT(value, ref.Type(), ""); nil == err {
			ref.Set(reflect.ValueOf(val))
		}
	}
	return
}

func (wr *bindWrapper) upAndContinue(key string, value interface{}) bool {
	if v, ok := wr.previousData[key]; ok {
		if nil == value && v == nil {
			return false // Same value
		}

		switch v2 := value.(type) {
		case string:
			if *v == v2 {
				return false
			}
		case []byte:
			if *v == string(v2) {
				return false
			}
		}
	}

	if nil == value {
		wr.previousData[key] = nil
	} else {
		var vv = new(string)
		switch v := value.(type) {
		case string:
			*vv = v
			wr.previousData[key] = vv
		case []byte:
			*vv = string(v)
			wr.previousData[key] = vv
		}
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func unpoint(ref reflect.Value) reflect.Value {
	for ref.IsValid() && ref.Kind() == reflect.Ptr && !ref.IsNil() {
		ref = ref.Elem()
	}
	return ref
}
