package netcommon

import (
	"context"
	"net"
	"syscall"
)

func ParaListen(listenAddress string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error

			err := conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
			})
			if err != nil {
				return err
			}
			return operr
		},
	}

	return lc.Listen(context.Background(), "tcp", listenAddress)
}

func ParaDialer() (dialer *net.Dialer) {
	dialer = &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			err := conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
			})
			if err != nil {
				return err
			}
			return operr
		},
	}
	return
}
