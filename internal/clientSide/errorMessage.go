package clientside

import (
	"houance/protoDemo-LoadBalance/internal/netCommon"
)
type ClientSideErrorMessage struct {
	Bes *netcommon.BasicErrorMessage
	StreamID uint32
}

func (csem *ClientSideErrorMessage) Error() string {
	return csem.Bes.Error()
}

func (csem *ClientSideErrorMessage) Wrap(bes *netcommon.BasicErrorMessage, streamID uint32)  {
	csem.Bes = bes
	csem.StreamID = streamID
}