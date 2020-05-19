package mustang

//import "fmt"
import log "github.com/Sirupsen/logrus"
import "github.com/hypebeast/go-osc/osc"

type OSC struct {
	Addr string
}

type OSCHandler func(*osc.Message)

func (o *OSC) Start(tx chan Message) {

	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/patch/name", func(msg *osc.Message) {
		osc.PrintMessage(msg)
		if len(msg.Arguments) > 0 {
			tx <- &PatchChange{
				Name: msg.Arguments[0].(string),
			}
		}
	})

	o.setupFxHandlers(d, tx)

	server := &osc.Server{
		Addr:       o.Addr,
		Dispatcher: d,
	}
	server.ListenAndServe()
}

func (o *OSC) ProcessMessages(rx chan Message) {
	for {
		msg := <-rx
		//handlers := msg.GetOSCHandlers()
		log.Info("message:", msg)
		/*for address, handler := range handlers {
			fmt.Printf("address=%v handler=%v\n", address, handler)
		}*/
	}
}

func (o *OSC) setupFxHandlers(d *osc.StandardDispatcher, tx chan Message) {
	handled := make(map[string]bool)
	for dsp, x := range fxParamNames {
		for model, y := range x {
			for paramIdx, name := range y {
				category := fxCategory[dsp]
				controlName := name
				log.Info("Setup ", category, model, paramIdx, " ", name)
				path := "/" + fxCategory[dsp] + "/" + name
				_, ok := handled[path]
				if ok {
					continue
				}
				log.Info(">> ", path)
				handled[path] = true
				_dsp := dsp
				d.AddMsgHandler(path, func(msg *osc.Message) {
					osc.PrintMessage(msg)
					if len(msg.Arguments) > 0 {
						value := uint32(msg.Arguments[0].(float32))
						tx <- &ParameterChange{
							Dsp:     _dsp,
							Control: controlName,
							Value:   value,
						}
					}
				})

			}
		}
	}
}

func (v *Version) GetOSCHandlers() (handlers map[string]OSCHandler) {
	return nil
}
func (v *ParameterChange) GetOSCHandlers() (handlers map[string]OSCHandler) {
	return nil
}
func (v *PatchChange) GetOSCHandlers() (handlers map[string]OSCHandler) {
	return nil
}
func (v *EffectChange) GetOSCHandlers() (handlers map[string]OSCHandler) {
	return nil
}
