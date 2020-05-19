package mustang

import log "github.com/Sirupsen/logrus"

func (p *Version) Send(m *Mustang)      {}
func (p *EffectChange) Send(m *Mustang) {}

func (p *ParameterChange) Send(m *Mustang) {
	log.Info("ParameterChange Send ", p.Dsp, p.Category, p.Control, p.Value)
	buf := make([]byte, 16)
	buf[0] = 0x05
	buf[1] = 0xc3

	if p.Dsp == 0 {
		for k, v := range fxCategory {
			if v == p.Category {
				p.Dsp = k
				break
			}
		}
	}

	if p.Dsp == 0 {
		return
	}

	buf[2] = byte(p.Dsp - 3) // 0x02

	model := m.CurrentEffect[p.Dsp].Model
	buf[3] = byte(model & 0xff)
	buf[4] = byte((model & 0xff00) >> 8)

	b, err := m.GetControlIdx(p.Dsp, model, p.Control)
	if err != nil {
		log.Warn("Could not lookup bytes for ", p.Dsp, p.Control)
		return
	}
	copy(buf[5:7], b)
	buf[7] = 0x0c

	buf[9] = 0
	buf[10] = byte(p.Value)

	m.UsbSend(buf)
}

func (p *PatchChange) Send(m *Mustang) {
	log.Info("PatchChange Send", p.Idx)
	buf := make([]byte, 64)
	buf[0] = 0x1c
	buf[1] = 0x01
	buf[2] = 0x01
	buf[3] = byte(p.Category & 0xff)
	buf[4] = byte(p.Idx & 0xff)
	buf[6] = 0x01
	// copy(buf[16:32], p.Name)

	m.UsbSend(buf)
}
