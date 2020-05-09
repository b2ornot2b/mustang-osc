package mustang

//import "fmt"
//import "log"
import "sort"
import "bytes"
import "errors"

import "encoding/binary"

type MustangUsbProtocolParser struct {
	buf []byte
}

type Message interface{}

type Version struct {
	Version string
}

type ParameterChange struct {
	Dsp     uint32
	Effect  uint32
	Control uint32
	Value   uint32
}

type PatchnameChange struct {
	Category uint32
	Idx      uint32
	Name     string
}

type Parameter struct {
	Name    string
	Control uint32
	Value   uint32
}

type EffectChange struct {
	Effect  string
	Name    string
	Model   uint32
	Enabled bool
	Params  []*Parameter
}

type paramNamesMap map[uint16](map[uint8]string)

func (p *MustangUsbProtocolParser) parseEffect(
	buf []byte,
	effect string,
	ids map[uint16]string,
	paramNames paramNamesMap) (*EffectChange, error) {
	model := binary.LittleEndian.Uint16(buf[16:18])
	name := ids[model]
	enabled := false
	if buf[22] == 0 {
		enabled = true
	}
	params := p.parseEffectParams(paramNames[model])
	m := EffectChange{
		Effect:  effect,
		Name:    name,
		Model:   uint32(model),
		Enabled: enabled,
		Params:  params,
	}
	return &m, nil
}

func (p *MustangUsbProtocolParser) parseEffectParams(paramNames map[uint8]string) []*Parameter {
	params := make([]*Parameter, 0, len(paramNames))
	for idx, paramName := range paramNames {
		value := p.buf[32+idx]
		params = append(params, &Parameter{
			Name:    paramName,
			Control: uint32(idx),
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
	case 0x0100:
		lastIdx := bytes.IndexByte(buf[4:], byte(0))
		version := string(buf[4 : 4+lastIdx])
		m := Version{Version: version}
		return &m, nil
	case 0x0501:
		dsp := uint32(buf[2])
		effect := uint32(buf[3])
		control := uint32(buf[5])
		value := uint32(binary.BigEndian.Uint16(buf[9:11]))
		if control >= 15 && control <= 20 {
			value >>= 8
		}
		m := ParameterChange{Dsp: dsp,
			Effect:  effect,
			Control: control,
			Value:   value,
		}
		return &m, nil
	case 0x1c01:
		dsp := uint32(buf[2])
		category := uint32(buf[3])
		idx := uint32(buf[4])
		lastIdx := bytes.IndexByte(buf[16:], byte(0))
		name := string(buf[16 : 16+lastIdx])
		switch dsp {
		case 0x04:
			m := PatchnameChange{
				Category: category,
				Idx:      idx,
				Name:     name,
			}
			return &m, nil
		case 0x05:
			model := binary.LittleEndian.Uint16(buf[16:18])
			name := ampIds[model]
			params := p.parseEffectParams(ampParamNames)
			m := EffectChange{
				Effect: "Amp",
				Name:   name,
				Model:  uint32(model),
				Params: params,
			}
			return &m, nil
		case 0x06:
			return p.parseEffect(buf, "Stompbox", stompboxIds, stompboxParamNames)
		case 0x07:
			return p.parseEffect(buf, "Modulation", modulationIds, modulationParamNames)
		case 0x08:
			return p.parseEffect(buf, "Delay", delayIds, delayParamNames)
		case 0x09:
			return p.parseEffect(buf, "Reverb", reverbIds, reverbParamNames)
		default:
			return nil, nil
		}

	}
	return nil, errors.New("Unknown message type")
}
