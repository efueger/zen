package zen

import (
	"log"
	"reflect"
	"testing"
)

func TestRouteStructSize(t *testing.T) {
	route := route{}
	log.Println(reflect.TypeOf(route).Size())
}
