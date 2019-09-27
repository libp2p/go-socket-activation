// +build linux
package sockets

import (
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"testing"
)

func TestActivation(t *testing.T) {
	lt, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv6loopback})
	if err != nil {
		t.Fatal(err)
	}
	if c, err := lt.SyscallConn(); err == nil {
		var fd uintptr
		c.Control(func(nfd uintptr) {
			fd = nfd
		})
		if fd != 3 {
			t.Fatal("expected fd to be 3")
		}
	} else {
		t.Fatal(err)
	}

	lu, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv6loopback})
	if err != nil {
		t.Fatal(err)
	}
	if c, err := lu.SyscallConn(); err == nil {
		var fd uintptr
		c.Control(func(nfd uintptr) {
			fd = nfd
		})
		if fd != 4 {
			t.Fatal("expected fd to be 3")
		}
	} else {
		t.Fatal(err)
	}

	os.Setenv("LISTEN_FDS", "2")
	os.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid()))
	os.Setenv("LISTEN_FDNAMES", "tcp:udp")
	tcp, err := TakeListeners("tcp")
	if err != nil {
		t.Fatal(err)
	}
	if len(tcp) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(tcp))
	}

	go func() {
		l := tcp[0]
		c, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		c.Write([]byte("foobar"))
		c.Close()
	}()
	c, err := net.DialTCP("tcp", nil, lt.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(c)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "foobar" {
		t.Error("wrong message")
	}
	udp, err := TakePacketConns("udp")
	if err != nil {
		t.Fatal(err)
	}
	if len(udp) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(tcp))
	}
	if udp[0].LocalAddr().String() != lu.LocalAddr().String() {
		t.Fatal("got the wrong udp listener")
	}
}
