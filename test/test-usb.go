package main

import log "github.com/Sirupsen/logrus"

import "time"
import "github.com/b2ornot2b/gomustang"

func main() {
	log.SetLevel(log.DebugLevel)

	fenderMustang := mustang.Mustang{}
	osc := mustang.OSC{"192.168.0.13:3000"}

	rx, tx := fenderMustang.GetChannels()
	log.Debug("rx: ", rx)
	go fenderMustang.Connect()

	go testSendParams(tx)
	go osc.Start(tx)
	go osc.ProcessMessages(rx)

	for {
		//msg := <-rx
		//log.Info("message:", msg)
		time.Sleep(100 * time.Second)
	}
}

func testSendParams(tx chan mustang.Message) {
	time.Sleep(2 * time.Second)
	log.Info("Sending...")
	tx <- &mustang.PatchChange{Idx: 13}

	time.Sleep(2 * time.Second)
	for {
		for i := 0; i <= 255; i++ {
			tx <- &mustang.ParameterChange{Category: "Amp", Control: "Gain", Value: uint32(i)}
			return
		}
		for i := 255; i >= 0; i-- {
			tx <- &mustang.ParameterChange{Category: "Amp", Control: "Gain", Value: uint32(i)}
		}

	}
	//tx <- pc
	//log.Info("Sent")

}
