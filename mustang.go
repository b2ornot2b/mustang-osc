package mustang

import "fmt"
import "time"
import "errors"
import "encoding/binary"
import log "github.com/Sirupsen/logrus"

import "github.com/google/gousb"

const ModelPosition = 16

type Controller interface{}

type Mustang struct {
	epIn    *gousb.InEndpoint
	epOut   *gousb.OutEndpoint
	channel []chan Message

	CurrentEffect map[DspType]EffectChange
}

func (m *Mustang) GetChannels() (rx chan Message, tx chan Message) {
	if m.channel == nil {
		m.channel = make([]chan Message, 2)
		m.channel[0] = make(chan Message)
		m.channel[1] = make(chan Message)

		go m.senderLoop(m.channel[1])
	}
	return m.channel[0], m.channel[1]
}

func (m *Mustang) Connect() {
	m.CurrentEffect = make(map[DspType]EffectChange)

	m.usbConnect()
}

func (m *Mustang) SendPatchname(name string) {
	log.Info("SendPatchname", name)

}

func (m *Mustang) GetControlIdx(dsp DspType, model FxModel, control string) ([]byte, error) {

	params := fxParamNames[dsp][model]

	for ctrlId, ctrlName := range params {
		if ctrlName == control {
			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(ctrlId))
			return b, nil
		}
	}
	return nil, errors.New("control not found")
}

func (m *Mustang) senderLoop(tx chan Message) {
	for {
		msg := <-tx
		log.Debug("usb tx: ", msg)
		msg.Send(m)
		time.Sleep(3 * time.Millisecond)
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

		p := MustangUsbProtocolParser{mustang: m}
		m.displayBuf(buf, readBytes)
		msg, err := p.parse(buf)
		if err != nil {
			log.Warn("Unhandled message...")
			//m.displayBuf(buf, readBytes)
			continue
		}

		if msg == nil {
			log.Debug("Silently dropping message...")
			//m.displayBuf(buf, readBytes)
			continue
		}

		//log.Debug("Dsp: ", msg.Dsp)

		//fmt.Printf("msg %t = %v\n", msg, msg)
		m.channel[0] <- msg
	}
}

func (m *Mustang) displayBuf(buf []byte, readBytes int) {
	return
	fmt.Printf("buf := []byte { ")
	for i := 0; i < readBytes; i++ {
		if true {
			fmt.Printf("\t%d: %02x %q", i, buf[i], buf[i])
			if 0 == ((i + 1) % 8) {
				fmt.Println()
			}
		} else {
			fmt.Printf("0x%02x, ", buf[i])
		}
	}
	fmt.Printf("}\n")
}

func (m *Mustang) usbInit() {
	buf := make([]byte, m.epOut.Desc.MaxPacketSize)
	buf[1] = 0xc3
	m.UsbSend(buf)

	buf[0] = 0x1a
	buf[1] = 0xc1
	m.UsbSend(buf)
}

func (m *Mustang) UsbSend(buf []byte) {
	epOut := m.epOut
	fmt.Println("UsbSend")
	m.displayBuf(buf, len(buf))
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
