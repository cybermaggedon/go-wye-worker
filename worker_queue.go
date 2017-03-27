
package wye

import (
	"os"
	"gopkg.in/redis.v5"
)

func getenv(env string, def string) string {
	s := os.Getenv(env)
	if s == "" {
		return def
	} else {
		return s
	}
}

type WorkerQueue struct {
	endpoints []string
	client   *redis.Client
	cur int
}

func (w *WorkerQueue) Send(msg []uint8) error {

	err := w.client.RPush(w.endpoints[w.cur], msg).Err()
	if err != nil {
		return err
	}

	w.cur = w.cur + 1
	if w.cur >= len(w.endpoints) {
		w.cur = 0
	}

	return nil
	
}

func NewWorkerQueue(endpoints []string) (*WorkerQueue, error) {

	w := new(WorkerQueue)

	w.client = redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_SERVER", "redis:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	w.endpoints = endpoints

	w.cur = 0

	return w, nil

}


