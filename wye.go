
package wye

import (
	"strings"
	"os"
	"fmt"
	"time"
	"gopkg.in/redis.v5"
	"github.com/google/uuid"
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
	client *redis.Client
	queue string
}

func (w *QueueWorker) CreateInput() string {

	u := uuid.New().String()
	return "q:" + u

}

type Handler interface {
	Handle(message []uint8, w *Worker) error
}

func (w *QueueWorker) Initialise(outputs []string) error {

	w.ctrl = os.NewFile(3, "fd3")

	_, err := w.ctrl.WriteString("INIT\n")
	if err != nil {
		return err
	}

	w.client = redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_SERVER", "localhost:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	w.out, err = w.ParseOutputs(outputs)
	if err != nil {
		return err
	}

	w.queue = w.CreateInput()
	if err != nil {
		return err
	}

	fmt.Fprintf(w.ctrl, "INPUT:input:%s\n", w.queue)
	if err != nil {
		return err
	}

	w.ctrl.WriteString("RUNNING\n")

	return nil

}

func (w *QueueWorker) Run(h Handler) {

	for {

		val, err := w.client.BLPop(0, w.queue).Result()

		if err == nil {
			h.Handle([]byte(val[1]), &(w.Worker))
			continue
		}

		if err == redis.Nil {
			continue
		}

		fmt.Println("Error: %s\n", err.Error())

		time.Sleep(1 * time.Second)

	}

}
