
package wye

import (
	"strings"
	"log"
	"os"
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

type EventHandler interface {
	Handle(message []uint8)
}

type Worker struct {
	ctrl *os.File
	out OutputSet
}

func (w *Worker) Initialise(outputs []string) {
	w.ctrl = os.NewFile(3, "fd3")
	w.ctrl.WriteString("INIT\n")
	w.out = w.ParseOutputs(outputs)
	w.ctrl.WriteString("RUNNING\n")
}

func (w *Worker) MakeOutput() OutputSet {
	a := OutputSet{}
	a.outputs = (make(map[string]*Output))
	return a
}

func (w *Worker) ParseOutputs(a []string) OutputSet {

	var outs OutputSet
	outs = w.MakeOutput()

	for _, elt := range a {

		toks := strings.SplitN(elt, ":", 2)

		name := toks[0]
		endpoints := strings.Split(toks[1], ",")

		outs.Add(name, endpoints)

	}

	return outs

}

func (w *Worker) Send(name string, msg []uint8) {
	w.out.outputs[name].Send(msg)
}

type QueueWorker struct {
	Worker
	in *zmq.Socket
	in_address string
}

func (w *QueueWorker) CreateInput(endpoint string) (*zmq.Socket, string) {

	var socket *zmq.Socket
	var err error
	socket, err = zmq.NewSocket(zmq.PULL)
	if err != nil {
		log.Fatalf("Couldn't create socket: %s", err.Error())
	}

	err = socket.Bind(endpoint)
	if err != nil {
		log.Fatalf("Couldn't bind socket: %s", err.Error())
	}
	
	addr, err := socket.GetLastEndpoint()
	if err != nil {
		log.Fatal("Couldn't get endpoint address")
	}
	
	return socket, addr

}

type Handler interface {
	Handle(message []uint8, w *Worker) error
}

func (w *QueueWorker) Receive(h Handler) error {
	msg, err := w.in.RecvBytes(0)
	if (err != nil) {
		return err
	}

	return h.Handle(msg, &(w.Worker))

}

func (w *QueueWorker) Initialise(outputs []string) {
	w.ctrl = os.NewFile(3, "fd3")
	w.ctrl.WriteString("INIT\n")
	w.out = w.ParseOutputs(outputs)
	w.in, w.in_address = w.CreateInput("tcp://*:*")
	fmt.Fprintf(w.ctrl, "INPUT:input:%s\n", w.in_address)
	w.ctrl.WriteString("RUNNING\n")
}

func (w *QueueWorker) Run(h Handler) {
	r := zmq.NewReactor()
	r.AddSocket(w.in, zmq.POLLIN,
		func(s zmq.State) error { return w.Receive(h) })
	r.Run(-1)
}
