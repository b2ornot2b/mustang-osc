package mustang

import "github.com/hypebeast/go-osc/osc"

type OSC struct {
	Addr string
}

func (o *OSC) Start() {

	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/message/address", func(msg *osc.Message) {
		osc.PrintMessage(msg)
	})

	server := &osc.Server{
		Addr:       o.Addr,
		Dispatcher: d,
	}
	server.ListenAndServe()
}
