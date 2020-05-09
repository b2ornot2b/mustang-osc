package main

import log "github.com/Sirupsen/logrus"

import "time"
import "github.com/b2ornot2b/gomustang"

func main() {
	fenderMustang := mustang.Mustang{}
	osc := mustang.OSC{"0.0.0.0:9000"}

	go osc.Start()

	rx, tx := fenderMustang.GetChannels()
	log.Debug("rx: ", rx)
	go fenderMustang.Connect()

	go testSendParams(tx)

	for {
		msg := <-rx
		log.Info("message:", msg)
		//time.Sleep(100 * time.Second)
	}
}

func testSendParams(tx chan mustang.Message) {
	time.Sleep(4 * time.Second)
	log.Info("Sending...")
	tx <- `{ "UpdateType": "UpdateAmp_PatchChange", "Idx": 13 }`
	log.Info("Sent")

}
