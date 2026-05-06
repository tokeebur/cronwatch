package notifier

import (
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// startFakeSMTP spins up a minimal TCP server that accepts one connection,
// reads the data sent to it, and returns it for inspection.
func startFakeSMTP(t *testing.T) (addr string, received func() string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	var buf strings.Builder
	done := make(chan struct{})

	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Minimal SMTP handshake expected by net/smtp.
		conn.Write([]byte("220 localhost ESMTP\r\n"))  //nolint:errcheck
		tmp := make([]byte, 4096)
		for {
			n, err := conn.Read(tmp)
			if n > 0 {
				buf.Write(tmp[:n])
				// Echo minimal SMTP responses.
				data := buf.String()
				switch {
				case strings.Contains(data, "EHLO"):
					conn.Write([]byte("250 OK\r\n")) //nolint:errcheck
				case strings.Contains(data, "MAIL FROM"):
					conn.Write([]byte("250 OK\r\n")) //nolint:errcheck
				case strings.Contains(data, "RCPT TO"):
					conn.Write([]byte("250 OK\r\n")) //nolint:errcheck
				case strings.Contains(data, "DATA"):
					conn.Write([]byte("354 Start\r\n")) //nolint:errcheck
				case strings.Contains(data, "\r\n.\r\n"):
					conn.Write([]byte("250 OK\r\n")) //nolint:errcheck
					return
				}
			}
			if err == io.EOF {
				return
			}
		}
	}()

	return ln.Addr().String(), func() string {
		<-done
		return buf.String()
	}
}

func TestSend_RendersJobName(t *testing.T) {
	addr, received := startFakeSMTP(t)

	host, port := splitHostPort(t, addr)
	n := New(Config{
		SMTPHost: host,
		SMTPPort: port,
		From:     "cronwatch@example.com",
		To:       []string{"ops@example.com"},
	})

	alert := Alert{
		JobName:   "backup-db",
		Command:   "/usr/bin/backup.sh",
		ExitCode:  1,
		Output:    "disk full",
		Err:       errors.New("exit status 1"),
		OccuredAt: time.Date(2024, 6, 1, 3, 0, 0, 0, time.UTC),
	}

	if err := n.Send(alert); err != nil {
		t.Fatalf("Send() unexpected error: %v", err)
	}

	body := received()
	if !strings.Contains(body, "backup-db") {
		t.Errorf("expected job name in email body, got:\n%s", body)
	}
	if !strings.Contains(body, "disk full") {
		t.Errorf("expected output in email body, got:\n%s", body)
	}
}

func splitHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return host, port
}
