
package wye

import (
	"log"
	zmq "github.com/pebbe/zmq4"
)

type Output struct {
	workers []WorkerQueue
}

func (o *Output) Add(endpoints []string) {

	w := WorkerQueue{}
	w.cur = 0

	for _, v := range(endpoints) {
		var socket *zmq.Socket
		var err error
		socket, err = zmq.NewSocket(zmq.PUSH)
		if err != nil {
			log.Fatalf("Couldn't create socket: %s", err.Error())
		}

		err = socket.Connect(v)
		if err != nil {
			log.Fatal("Couldn't connect")
		}

		w.sockets = append(w.sockets, socket)
	}
	
	o.workers = append(o.workers, w)

}

func (o *Output) Send(msg []uint8) {
	for _, w := range o.workers {
		w.Send(msg)
	}
}
