# connproxy
[![Go Reference](https://pkg.go.dev/badge/github.com/asciimoth/connproxy.svg)](https://pkg.go.dev/github.com/asciimoth/connproxy)  

A simple net.Listener <-> net.Conn proxy

## Example
Launch this code and run `curl 127.0.0.1:8080`
```go
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/asciimoth/connproxy"
)

func main() {
	listener, err := net.Listen("tcp4", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Println("Starting TCP proxy")
	connproxy.Proxy{
		Accept: listener.Accept,
		Dial: func(_ context.Context) net.Conn {
			fmt.Println("New outgoing conn")
			conn, err := net.Dial("tcp4", "google.com:http")
			if err != nil {
				return nil
			}
			return conn
		},
	}.Run(ctx)
	fmt.Println("Proxy stopped")
}
```


## License
Files in this repository are distributed under the CC0 license.  

<p xmlns:dct="http://purl.org/dc/terms/">
  <a rel="license"
     href="http://creativecommons.org/publicdomain/zero/1.0/">
    <img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />
  </a>
  <br />
  To the extent possible under law,
  <a rel="dct:publisher"
     href="https://github.com/asciimoth">
    <span property="dct:title">ASCIIMoth</span></a>
  has waived all copyright and related or neighboring rights to
  <span property="dct:title">connproxy</span>.
</p>

