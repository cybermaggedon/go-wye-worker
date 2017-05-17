
package wye

import (
	"os"
	"gopkg.in/redis.v5"
	"time"
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
	cur int
	internalQueue chan []uint8
}

func (w *WorkerQueue) qWriter() {

	client := redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_SERVER", "redis:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()

	for {

		msg := <- w.internalQueue

		for {

			err := client.RPush(w.endpoints[w.cur], msg).Err()
			if err == nil {
				break
			}

			// Redis error, reconnect and retry.
			client.Close()

			time.Sleep(5 * time.Second)

			client = redis.NewClient(&redis.Options{
				Addr:     getenv("REDIS_SERVER", "redis:6379"),
				Password: "", // no password set
				DB:       0,  // use default DB
			})

			w.cur = w.cur + 1
			if w.cur >= len(w.endpoints) {
				w.cur = 0
			}
			
			continue

		}
				
	}

}

func (w *WorkerQueue) Send(msg []uint8) error {

	w.internalQueue <- msg

	return nil
	
}

func NewWorkerQueue(endpoints []string) (*WorkerQueue, error) {

	w := new(WorkerQueue)

	w.internalQueue = make(chan []uint8, 50)

	w.endpoints = endpoints

	w.cur = 0

	go w.qWriter()

	return w, nil

}


