
package wye

type OutputSet struct {
	outputs map[string]*Output
}

func NewOutputSet() *OutputSet {
	s := &OutputSet{}
	s.outputs = make(map[string]*Output)
	return s
}

func (o *OutputSet) Add(name string, endpoints []string) error {
	if _, ok := o.outputs[name]; !ok {
		o.outputs[name] = &(Output{})
	}

	return o.outputs[name].Add(endpoints)

}

func (o *OutputSet) Send(name string, msg []uint8) error {
	return o.outputs[name].Send(msg)
}

