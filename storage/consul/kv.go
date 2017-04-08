//
// @project registry 2017
// @author Dmitry Ponomarev <demdxx@gmail.com> 2017
//

package consul

import (
	"github.com/hashicorp/consul/api"
)

type kv struct {
	prefix string
	client *api.KV
}

// Get value
func (kv kv) Get(key string) (string, error) {
	v, _, err := kv.client.Get(kv.key(key), nil)
	if err != nil {
		return "", err
	}
	if v != nil {
		return string(v.Value), nil
	}
	return "", nil
}

// Set value for key
func (kv kv) Set(key, value string) (err error) {
	_, err = kv.client.Put(&api.KVPair{
		Key:   kv.key(key),
		Value: []byte(value),
	}, nil)
	return err
}

// List of valuse
func (kv kv) List(prefix string) (map[string]string, error) {
	var items, _, err = kv.client.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	var list = make(map[string]string, len(items))
	for _, k := range items {
		list[k.Key] = string(k.Value)
	}

	return list, nil
}

// Delete item by key
func (kv kv) Delete(key string) (err error) {
	_, err = kv.client.Delete(kv.key(key), nil)
	return
}

func (kv kv) key(k string) string {
	if kv.prefix != "" {
		return kv.prefix + "/" + k
	}
	return k
}
