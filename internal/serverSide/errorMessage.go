package serverside

type ServerSideErrorMessage struct {
	address string
	errString string
	err error
}

func (ssem *ServerSideErrorMessage) Error() string {
	return ssem.err.Error()
}

func (ssem *ServerSideErrorMessage) Wrap(address string ,err error)  {
	ssem.address = address
	ssem.err = err
}

func (ssem *ServerSideErrorMessage) GetAddress() string  {
	return ssem.address
}