
package wye

import (
	zmq "github.com/pebbe/zmq4"
)

type WorkerQueue struct {
	sockets []*zmq.Socket
	cur int
}

func (w *WorkerQueue) Send(msg []uint8) error {

	_, err := w.sockets[w.cur].SendBytes(msg, 0)
	if err != nil {
		return err
	}

	w.cur = w.cur + 1
	if w.cur >= len(w.sockets) {
		w.cur = 0
	}

	return nil
	
}

func NewWorkerQueue(endpoints []string) (*WorkerQueue, error) {

	w := new(WorkerQueue)

	for _, v := range(endpoints) {
		var socket *zmq.Socket
		var err error

		socket, err = zmq.NewSocket(zmq.PUSH)
		if err != nil {
			return nil, err
		}

		err = socket.Connect(v)
		if err != nil {
			return nil, err
		}

		w.sockets = append(w.sockets, socket)
	}

	w.cur = 0

	return w, nil

}


