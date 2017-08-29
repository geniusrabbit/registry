package docker

import (
	"fmt"
	"testing"
)

func TestIP(t *testing.T) {
	fmt.Println(resolveLocalIP())
}
