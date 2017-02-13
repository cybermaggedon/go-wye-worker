
package wye

import (
	zmq "github.com/pebbe/zmq4"
)

type WorkerQueue struct {
	sockets []*zmq.Socket
	cur int
}

func (w *WorkerQueue) Send(msg []uint8) {
	w.sockets[w.cur].SendBytes(msg, 0)
	w.cur++
	if w.cur > len(w.sockets) {
		w.cur = 0
	}
}

