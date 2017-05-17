
package wye

import (
	"strings"
	"os"
	"fmt"
	"time"
	"gopkg.in/redis.v5"
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
	queue string
}

type Handler interface {
	Handle(message []uint8, w *Worker) error
}

func (w *QueueWorker) Initialise(input string, outputs []string) error {

	w.ctrl = os.NewFile(3, "fd3")

	_, err := w.ctrl.WriteString("INIT\n")
	if err != nil {
		return err
	}

	w.out, err = w.ParseOutputs(outputs)
	if err != nil {
		return err
	}

	w.queue = "q:" + input

	fmt.Fprintf(w.ctrl, "INPUT:input:%s\n", w.queue)
	if err != nil {
		return err
	}

	w.ctrl.WriteString("RUNNING\n")

	return nil

}

func (w *QueueWorker) qReader(ch chan []uint8) {

	client := redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_SERVER", "redis:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for {

		val, err := client.BLPop(0, w.queue).Result()

		if err == nil {
			bts := []byte(val[1])
			ch <- bts
			continue
		}

		// No message.  It's blocking, so...
		// FIXME: Case doesn't happen?
		if err == redis.Nil {
			continue
		}

		client.Close()

		time.Sleep(5 * time.Second)

		client = redis.NewClient(&redis.Options{
			Addr:     getenv("REDIS_SERVER", "redis:6379"),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

	}

}

func (w *QueueWorker) Run(h Handler) {

	ch := make(chan []uint8, 10)

	go w.qReader(ch)

	for {
		val := <- ch
		h.Handle(val, &(w.Worker))
	}

}
