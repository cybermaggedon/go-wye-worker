
package wye

import (
	"strings"
	"os"
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

type EventHandler interface {
	Handle(message []uint8)
}

type Worker struct {
	ctrl *os.File
	out *OutputSet
}

func (w *Worker) Initialise(outputs []string) error {

	w.ctrl = os.NewFile(3, "fd3")

	_, err := w.ctrl.WriteString("INIT\n")
	if err != nil {
		return err
	}

	w.out, err = w.ParseOutputs(outputs)
	if err != nil {
		return err
	}

	_, err = w.ctrl.WriteString("RUNNING\n")
	if err != nil {
		return err
	}

	return nil

}

func (w *Worker) ParseOutputs(a []string) (*OutputSet, error) {

	outs := NewOutputSet()

	for _, elt := range a {

		toks := strings.SplitN(elt, ":", 2)

		name := toks[0]
		endpoints := strings.Split(toks[1], ",")

		err := outs.Add(name, endpoints)
		if err != nil {
			return nil, err
		}

	}

	return outs, nil

}

func (w *Worker) Send(name string, msg []uint8) {
	w.out.outputs[name].Send(msg)
}

type QueueWorker struct {
	Worker
	in *zmq.Socket
	in_address string
}

func (w *QueueWorker) CreateInput(endpoint string) (*zmq.Socket, string, error) {

	var socket *zmq.Socket
	var err error
	socket, err = zmq.NewSocket(zmq.PULL)
	if err != nil {
		return nil, "", err
	}

	err = socket.Bind(endpoint)
	if err != nil {
		return nil, "", err
	}
	
	addr, err := socket.GetLastEndpoint()
	if err != nil {
		return nil, "", err
	}
	
	return socket, addr, nil

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

func (w *QueueWorker) Initialise(outputs []string) error {

	w.ctrl = os.NewFile(3, "fd3")

	_, err := w.ctrl.WriteString("INIT\n")
	if err != nil {
		return err
	}

	w.out, err = w.ParseOutputs(outputs)
	if err != nil {
		return err
	}

	w.in, w.in_address, err = w.CreateInput("tcp://*:*")
	if err != nil {
		return err
	}

	fmt.Fprintf(w.ctrl, "INPUT:input:%s\n", w.in_address)
	if err != nil {
		return err
	}

	w.ctrl.WriteString("RUNNING\n")

	return nil

}

func (w *QueueWorker) Run(h Handler) {
	r := zmq.NewReactor()
	r.AddSocket(w.in, zmq.POLLIN,
		func(s zmq.State) error { return w.Receive(h) })
	r.Run(-1)
}
