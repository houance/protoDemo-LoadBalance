package netcommon

import binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"

type BasicErrorMessage struct {
	Framer *binaryframer.BinaryFramer
	Err error
}

func (bes *BasicErrorMessage) Error() string {
	return bes.Err.Error()
}

func (bes *BasicErrorMessage) Wrap(framer *binaryframer.BinaryFramer, err error) {
	bes.Framer = framer
	bes.Err = err
}
