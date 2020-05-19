package mustang

//import "fmt"
//import "log"
import "sort"
import "bytes"
import "errors"

import "encoding/binary"

type MustangUsbProtocolParser struct {
	buf     []byte
	mustang *Mustang
}

type Message interface {
	Send(*Mustang)
}

type DspType uint16
type FxModel uint16
type ParameterIdx uint8

type Version struct {
	Version string `/version`
}

type PatchChange struct {
	Category uint32
	Idx      uint32
	Name     string
	buf      []byte
}

type Parameter struct {
	Control ParameterIdx
	Value   uint32

	Name string
}

type EffectChange struct {
	Dsp   DspType
	Model FxModel

	Params  []*Parameter
	Enabled bool

	Category string
	Name     string

	buf []byte
}

type ParameterChange struct {
	Dsp   DspType
	Model FxModel

	ControlIdx ParameterIdx
	Value      uint32

	Effect   string
	Category string
	Control  string
	buf      []byte
}

type paramNamesMap map[FxModel](map[ParameterIdx]string)

func (p *MustangUsbProtocolParser) parseEffect(
	buf []byte,
	dsp DspType,
	category string,
	ids map[FxModel]string,
	paramNames paramNamesMap) (*EffectChange, error) {
	model := FxModel(binary.LittleEndian.Uint16(buf[16:18]))
	name := ids[model]
	enabled := false
	if buf[22] == 0 {
		enabled = true
	}
	params := p.parseEffectParams(paramNames[model])
	m := EffectChange{
		Dsp:   dsp,
		Model: model,

		Params:  params,
		Enabled: enabled,

		Category: category,
		Name:     name,

		buf: buf,
	}
	p.mustang.CurrentEffect[dsp] = m
	return &m, nil
}

func (p *MustangUsbProtocolParser) parseEffectParams(paramNames map[ParameterIdx]string) []*Parameter {
	params := make([]*Parameter, 0, len(paramNames))
	for idx, paramName := range paramNames {
		value := p.buf[32+idx]
		params = append(params, &Parameter{
			Name:    paramName,
			Control: idx,
			Value:   uint32(value),
		})
	}
	sort.Slice(params, func(i, j int) bool {
		return params[i].Control < params[j].Control
	})
	return params
}

func (p *MustangUsbProtocolParser) parse(buf []byte) (Message, error) {
	p.buf = buf
	switch binary.BigEndian.Uint16(buf[0:2]) {
	case 0x0000: // Initialization Response
		return nil, nil
	case 0x0100: // Initialization Version
		lastIdx := bytes.IndexByte(buf[4:], byte(0))
		version := string(buf[4 : 4+lastIdx])
		m := Version{Version: version}
		return &m, nil
	case 0x0501: // Parameter Change
		dsp := DspType(buf[2] + 3)
		effect := FxModel(binary.LittleEndian.Uint16(buf[3:5]))
		control := ParameterIdx(buf[5])
		value := uint32(binary.BigEndian.Uint16(buf[9:11]))
		if control >= 15 && control <= 20 {
			value >>= 8
		}
		name := fxParamNames[dsp][effect][control]
		m := ParameterChange{Dsp: dsp,
			Model:      effect,
			ControlIdx: control,
			Value:      value,

			Category: fxCategory[dsp],
			Effect:   fxNames[dsp][effect],
			Control:  name,

			buf: buf,
		}
		return &m, nil
	case 0x1c01: // Patch Change
		dsp := DspType(buf[2])
		category := uint32(buf[3])
		idx := uint32(buf[4])
		lastIdx := bytes.IndexByte(buf[16:], byte(0))
		name := string(buf[16 : 16+lastIdx])
		switch dsp {
		case 0x04:
			m := PatchChange{
				Category: category,
				Idx:      idx,
				Name:     name,
				buf:      buf,
			}
			return &m, nil
		case 0x05: // Amp Change
			return p.parseEffect(buf, dsp, "Amp", ampIds, ampParamNames)
		case 0x06: // Stompbox Change
			return p.parseEffect(buf, dsp, "Stompbox", stompboxIds, stompboxParamNames)
		case 0x07: // Modulation Change
			return p.parseEffect(buf, dsp, "Modulation", modulationIds, modulationParamNames)
		case 0x08: // Delay Change
			return p.parseEffect(buf, dsp, "Delay", delayIds, delayParamNames)
		case 0x09: // Reverb Change
			return p.parseEffect(buf, dsp, "Reverb", reverbIds, reverbParamNames)
		default:
			return nil, nil
		}

	}
	return nil, errors.New("Unknown message type")
}
