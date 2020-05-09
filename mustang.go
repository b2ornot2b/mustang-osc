package mustang

import "fmt"
import log "github.com/Sirupsen/logrus"

import "github.com/google/gousb"

const ModelPosition = 16

type Mustang struct {
	epIn      *gousb.InEndpoint
	epOut     *gousb.OutEndpoint
	jsChannel []chan string
	Channel   []chan string
}

func (m *Mustang) GetJsonChannels() (rx chan string, tx chan string) {
	if m.jsChannel == nil {
		m.jsChannel = make([]chan string, 2)
		m.jsChannel[0] = make(chan string)
		m.jsChannel[1] = make(chan string)

		go m.senderLoop(m.jsChannel[1])
	}
	return m.jsChannel[0], m.jsChannel[1]
}

func (m *Mustang) Connect() {
	m.usbConnect()
}

func (m *Mustang) senderLoop(tx chan string) {
	for {
		msg := <-tx
		log.Warn("Received: ", msg)
		/*
			m := &UpdateAmp{}
			err := protojson.Unmarshal([]byte(msg), m)
			if err != nil {
				log.Error("Invalid JSON", err)
				continue
			}
		*/

	}
}

func (m *Mustang) usbReaderLoop() {
	buf := make([]byte, m.epIn.Desc.MaxPacketSize)

	stream, err := m.epIn.NewStream(m.epIn.Desc.MaxPacketSize, 1)
	if err != nil {
		log.Fatal("NewStream failed", err)
	}

	for {
		readBytes, err := stream.Read(buf)
		if err != nil {
			log.Error("stream read failed", err)
			return
		}

		p := MustangUsbProtocolParser{}
		// m.displayBuf(buf, readBytes)
		msg, err := p.parse(buf)
		if err != nil {
			log.Warn("Unhandled message...")
			m.displayBuf(buf, readBytes)
			continue
		}

		if msg == nil {
			log.Debug("Silently dropping message...")
			continue
		}

		fmt.Printf("msg %t = %v\n", msg, msg)
	}
}

func (m *Mustang) displayBuf(buf []byte, readBytes int) {
	for i := 0; i < readBytes; i++ {
		fmt.Printf("\t%d: %02x %q", i, buf[i], buf[i])
		if 0 == ((i + 1) % 8) {
			fmt.Println()
		}
	}
}

func (m *Mustang) usbInit() {
	buf := make([]byte, m.epOut.Desc.MaxPacketSize)
	buf[1] = 0xc3
	m.usbSend(buf)

	buf[0] = 0x1a
	buf[1] = 0xc1
	m.usbSend(buf)
}

func (m *Mustang) usbSend(buf []byte) {
	epOut := m.epOut
	for retries := 0; retries < 64; retries++ {
		writeBytes, err := epOut.Write(buf)
		if err != nil {
			fmt.Println("Write returned an error:", err)
			continue
		}

		log.Debug("wrote bytes", writeBytes)
		break
	}
}

func (m *Mustang) usbConnect() {
	ctx := gousb.NewContext()
	defer ctx.Close()
	vid, pid := gousb.ID(0x1ed8), gousb.ID(0x0016)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == vid && desc.Product == pid
	})
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		log.Fatalf("OpenDevices(): %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("no devices found matching VID %s and PID %s", vid, pid)
	}

	dev := devs[0]

	err = dev.SetAutoDetach(true)
	if err != nil {
		log.Fatalf("setting auto detach failed %v", err)
	}

	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()
	defer intf.Close()

	epIn, err := intf.InEndpoint(1)
	if err != nil {
		log.Fatalf("%s.InEndpoint(1): %v", intf, err)
	}
	m.epIn = epIn

	epOut, err := intf.OutEndpoint(1)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(1: %v", intf, err)
	}
	m.epOut = epOut

	m.usbInit()
	m.usbReaderLoop()
}
