// +build linux

package sockets

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestActivation(t *testing.T) {
	testTcpAddr := os.Getenv("SOCKET_TEST_TCP_ADDR")
	testUdpAddr := os.Getenv("SOCKET_TEST_UDP_ADDR")
	if testTcpAddr == "" {
		lt, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv6loopback})
		if err != nil {
			t.Fatal(err)
		}
		tcpAddr := lt.Addr()
		ltf, err := lt.File()
		lt.Close()
		if err != nil {
			t.Fatal(err)
		}

		lu, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv6loopback})
		if err != nil {
			t.Fatal(err)
		}
		udpAddr := lu.LocalAddr()
		luf, err := lu.File()
		lu.Close()
		if err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command("/proc/self/exe", "-test.run", "TestActivation")
		cmd.ExtraFiles = append(cmd.ExtraFiles, ltf, luf)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(
			cmd.Env,
			"LISTEN_FDS=2",
			"LISTEN_FDNAMES=tcp:udp",
			fmt.Sprintf("SOCKET_TEST_TCP_ADDR=%s", tcpAddr),
			fmt.Sprintf("SOCKET_TEST_UDP_ADDR=%s", udpAddr),
		)
		err = cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	} else {
		// can't set this from the parent because we can't run anything between fork/exec.
		os.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid()))
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
		c, err := net.Dial("tcp6", testTcpAddr)
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
		if udp[0].LocalAddr().String() != testUdpAddr {
			t.Fatal("got the wrong udp listener")
		}
	}
}
