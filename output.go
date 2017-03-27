
package wye

type Output struct {
	workers []*WorkerQueue
}

func (o *Output) Add(endpoints []string) error {

	w, err := NewWorkerQueue(endpoints)
	if err != nil {
		return err
	}

	o.workers = append(o.workers, w)

	return nil

}

func (o *Output) Send(msg []uint8) error {
	for _, w := range o.workers {
		err := w.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}
