// Package connproxy provides simple net.Listener <-> net.Conn proxy
package connproxy

import (
	"context"
	"io"
	"net"
	"sync"
)

// Proxy pairs connections incoming from listener to outgoing
// connections created via dialler.
type Proxy struct {
	// Accept returns new incoming conn or error
	Accept func() (net.Conn, error)
	// Dial returns new outgoing conn or nil
	Dial func(context.Context) net.Conn
}

// Run accepts incoming connections via Proxy.Accept, pairs each with an
// outgoing connection created by Proxy.Dial, and relays data between each
// pair in its own goroutine.
// Run returns when ctx is canceled or when Proxy.Accept returns an error.
//
// Behavior notes:
//   - If p.Accept is nil, Run returns immediately.
//   - If p.Dial is nil or returns nil for a particular accept, the incoming
//     connection is closed and Run continues accepting others.
//   - When Run finishes (ctx canceled or Accept error) it closes all currently
//     managed connections and waits for handler goroutines to exit.
func (p Proxy) Run(ctx context.Context) {
	if p.Accept == nil {
		return
	}

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		next  uint64 // next conn id
		conns = make(map[uint64]net.Conn)
	)

	removeConn := func(id uint64) {
		mu.Lock()
		c := conns[id]
		if c != nil {
			_ = c.Close()
		}
		delete(conns, id)
		mu.Unlock()
	}

	defer func() {
		// close snapshot of connections (avoid holding lock while Close runs)
		mu.Lock()
		list := make([]net.Conn, 0, len(conns))
		for _, c := range conns {
			list = append(list, c)
		}
		mu.Unlock()
		for _, c := range list {
			_ = c.Close()
		}
		// wait for all handler goroutines to exit
		wg.Wait()
	}()

	type acceptResult struct {
		conn net.Conn
		err  error
	}

	for {
		// fast-path: exit if context already canceled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// run Accept in a goroutine so we can also select on ctx cancellation.
		// buffer size 1 prevents goroutine leak on immediate return.
		acceptCh := make(chan acceptResult, 1)
		go func() {
			c, err := p.Accept()
			acceptCh <- acceptResult{conn: c, err: err}
		}()

		select {
		case <-ctx.Done():
			return
		case res := <-acceptCh:
			if res.err != nil {
				return
			}
			in := res.conn
			if in == nil {
				// nothing to do
				continue
			}

			// create outgoing connection (if Dial is provided)
			var out net.Conn
			if p.Dial != nil {
				out = p.Dial(ctx)
			}
			if out == nil {
				// no outgoing side — close incoming and continue
				_ = in.Close()
				continue
			}

			inID := next
			outID := next + 1
			next += 2

			// Register
			conns[inID] = in
			conns[outID] = out
			wg.Go(func() {
				defer removeConn(inID)
				defer removeConn(outID)
				_, _ = io.Copy(out, in) // in -> out
			})
			wg.Go(func() {
				defer removeConn(inID)
				defer removeConn(outID)
				_, _ = io.Copy(in, out) // out -> in
			})
		}
	}
}
