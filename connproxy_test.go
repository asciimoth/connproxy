package connproxy_test

import (
	"github.com/asciimoth/connproxy"
	"testing"
)

func Test_GetHello(t *testing.T) {
	hello := connproxy.GetHello("World")
	if hello != "Hello World" {
		t.Fatalf("GetHello(\"World\") == %s", hello)
	}
}
