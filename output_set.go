
package wye

type OutputSet struct {
	outputs map[string]*Output
}

func (o *OutputSet) Add(name string, endpoints []string) {

	if _, ok := o.outputs[name]; !ok {
		o.outputs[name] = &(Output{})
	}

	o.outputs[name].Add(endpoints)

}

func (o *OutputSet) Send(name string, msg []uint8) {
	o.outputs[name].Send(msg)
}

