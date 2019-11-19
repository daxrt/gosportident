package sireader

import (
	"github.com/pkg/errors"
	"log"
	"time"

	"github.com/tarm/serial"
)

//ProtoConfig is configuration struct
type ProtoConfig struct {
	ExtProto  byte
	AutoSend  byte
	HandShake byte
	PWAccess  byte
	PunchRead byte
}

//Reader is main struct
type Reader struct {
	port        *serial.Port
	protoConfig ProtoConfig
	debug       bool
	logfile     string
}

//NewReader is constructor of Reader
func NewReader(port string) (*Reader, error) {
	c := &serial.Config{Name: port, Baud: 38400, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	r := new(Reader)
	r.port = s
	//flush possibly available input
	err = r.port.Flush()
	if err != nil {
		return nil, err
	}
	return r, nil
}

//func (r *Reader) SetExtendedProtocol(extended bool) {}

//func (r *Reader) SetAutoSend(autoSend bool) {}

//func (r *Reader) SetOperatingMode(mode string) {}

//func (r *Reader) SetStationCode(code int) {}

//GetTime is returning time object
func (r *Reader) GetTime() *time.Time {
	_, _, parameters, err := r.sendCommand([]byte{CGetTime}, []byte{})
	if err != nil {
		return nil
	}
	year := toInt(parameters[:1]) + 1971
	month := toInt(parameters[1:2])
	day := toInt(parameters[2:3])
	amPm := toInt(parameters[3:4]) & 0x1
	second := toInt(parameters[4:6])
	hour := amPm*12 + second/3600
	second %= 3600
	minute := second / 60
	second %= 60
	ms := toInt(parameters[6:7]) / 256.0 * 1000000
	res := time.Date(year, time.Month(month+1), day, hour, minute, second, ms*1000000, time.Local)
	return &res
}

//func (r *Reader) SetTime(t *time.Time) {}

//Beep cmd for si station
func (r *Reader) Beep() {
	_, _, _, _ = r.sendCommand([]byte{CBeep}, toBytes(1))
}

//func (r *Reader) PowerOff() {}

//Disconnect is closing port
func (r *Reader) Disconnect() error {
	return r.port.Close()
}

//func (r *Reader) Reconnect() {}

func (r *Reader) updateProtoConfig() ProtoConfig {
	//ret, _ := r.sendCommand(Bytes(C_GET_SYS_VAL),Bytes(O_PROTO, 0x01))
	//configByte := uint(toInt(Bytes()))
	protoConfig := ProtoConfig{}
	//protoConfig.ExtProto = configByte & (1 << 0) != 0
	//protoConfig.AutoSend = configByte & (1 << 1) != 0
	//protoConfig.HandShake = configByte & (1 << 2) != 0
	//protoConfig.PWAccess = configByte & (1 << 4) != 0
	//protoConfig.PunchRead = configByte & (1 << 7) != 0
	return protoConfig
}

//func (r *Reader) SetProtoConfig(config ProtoConfig) {}

func decodeCardNr(number int) int {
	return 0
}

//func decodeTime() {}

//func appendPunch() {}

//func decodeCardData() {}

func (r *Reader) sendCommand(command, parameters []byte) ([]byte, int, []byte, error) {
	cmd := BytesMerge(command, toBytes(len(parameters)), parameters)
	cmd = BytesMerge([]byte{STX}, cmd, crc(cmd), []byte{ETX})

	_, err := r.port.Write(cmd)
	if err != nil {
		return nil, 0, nil, err
	}
	return r.readCommand()
}

func (r *Reader) readCommand() ([]byte, int, []byte, error) {
	var cmd []byte
	var parameters []byte
	var crc []byte
	var stationCode []byte
	var parametersLength int
	var i int
Loop:
	for {
		buf := make([]byte, 128)
		n, err := r.port.Read(buf)
		if err != nil {
			return nil, 0, nil, err
		}
		log.Printf("buf %q", buf[:n])
		for _, b := range buf[:n] {
			if i == 1 {
				cmd = []byte{b}
			}
			if i == 2 {
				parametersLength = toInt([]byte{b})
			}
			if i > 2 && i <= 4 {
				stationCode = append(stationCode, b)
			}
			if i > 4 && i <= parametersLength+2 {
				parameters = append(parameters, b)
			}
			if i > parametersLength+2 && i <= parametersLength+4 {
				crc = append(crc, b)
			}
			if i == parametersLength+5 {
				break Loop
			}
			i++
		}

	}
	log.Printf("cmd %q station %q parameters %q crc %q", cmd, stationCode, parameters, crc)
	if !crcCheck(BytesMerge(cmd, toBytes(parametersLength), stationCode, parameters), crc) {
		return nil, 0, nil, errors.New("CRC check failed")
	}
	return cmd, toInt(stationCode), parameters, nil
}

//ReaderReadout is struct for readout
type ReaderReadout struct {
	Reader
	card     string
	CardType string
}

//func (rr *ReaderReadout) PollCard() {}

//func (rr *ReaderReadout) ReadCard(refTime int) {}

//func (rr *ReaderReadout) AckCard() {}

//func (rr *ReaderReadout) readCommand(timeout int) {}

//ReaderControl is struct for control
type ReaderControl struct {
	Reader
	nextOffset string
}

//func (rc *ReaderControl) PolPunch() {}

//func (rc *ReaderControl) readPunch(offset int) {}
