package serverside

import (
	"houance/protoDemo-LoadBalance/internal/netCommon"
)

type ServerSideErrorMessage struct {
	Bes *netcommon.BasicErrorMessage
}

func (ssem *ServerSideErrorMessage) Error() string {
	return ssem.Bes.Error()
}