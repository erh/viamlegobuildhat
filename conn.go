package viambuildhat

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"go.uber.org/multierr"
)

//go:embed data/firmware.bin
var firmware []byte

//go:embed data/signature.bin
var signature []byte

// path is usually /dev/serial0
func NewConnection(path string) (*Connection, error) {

	options := serial.OpenOptions{
		PortName:        path,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	
	dev, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	conn := &Connection{dev: dev}
	conn.smallSleep()

	go conn.readLoop()
	conn.smallSleep()

	
	err = conn.write([]byte("\r\rversion\r"))
	if err != nil {
		return nil, multierr.Combine(conn.Close(), err)
	}

	time.Sleep(time.Second)

	if false {
		err = conn.loadFirmware()
		if err != nil {
			return nil, multierr.Combine(conn.Close(), err)
		}
	}
	
	if true {
		err = conn.write([]byte("echo 0\r"))
		if err != nil {
			panic(err)
		}
		
		err = conn.write([]byte("port 3; plimit 0.7; select 0; pwm; set .8\r"))
		if err != nil {
			panic(err)
		}
		
		time.Sleep(time.Second*2)
		
		err = conn.write([]byte("port 3; select 0; pwm; set 0\r"))
		if err != nil {
			panic(err)
		}
	}

	
	return conn, nil
}


type Connection struct {
	dev io.ReadWriteCloser
}

func (c *Connection) readLoop() {
	buf := bufio.NewReader(c.dev)
	for {
		line, err := buf.ReadString('\r')
		if err != nil {
			if err == io.ErrClosedPipe {
				panic("TODO - ErrClosedPipe - should i reconnect")
			}
			panic(err)
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		
		fmt.Printf("got line [%s]\n", line)
	}
}

func (c *Connection) loadFirmware() error {
	err := c.write([]byte("clear\r"))
	if err != nil {
		return err
	}

	c.smallSleep() // TODO: self.getprompt()

	err = c.write([]byte(fmt.Sprintf("load %d %d\r", len(firmware), checksum(firmware))))
	if err != nil {
		return err
	}

	c.smallSleep()

	err = c.writeBinaryFile(firmware)
	if err != nil {
		return err
	}

	c.smallSleep() // TODO: self.getprompt()	

	err = c.write([]byte(fmt.Sprintf("signature %d\r", len(signature))))
	if err != nil {
		return err
	}

	c.smallSleep()
	
	err = c.writeBinaryFile(signature)
	if err != nil {
		return err
	}

	c.smallSleep() // TODO: self.getprompt()	

	c.write([]byte("reboot\r"))
	time.Sleep(time.Second)
	
	return nil
}

func (c *Connection) Close() error {
	/*
	               self.fin = True
            self.running = False
            self.th.join()
            self.cbqueue.put(())
            for q in self.motorqueue:
                q.put((None, None))
            self.cb.join()
            turnoff = ""
            for p in range(4):
                conn = self.connections[p]
                if conn.typeid != 64:
                    turnoff += f"port {p} ; pwm ; coast ; off ;"
                else:
                    hexstr = ' '.join(f'{h:x}' for h in [0xc2, 0, 0, 0, 0, 0, 0, 0, 0, 0])
                    self.write(f"port {p} ; write1 {hexstr}\r".encode())
            self.write(f"{turnoff}\r".encode())
            self.write(b"port 0 ; select ; port 1 ; select ; port 2 ; select ; port 3 ; select ; echo 0\r")

	*/
	return c.dev.Close()
}

func (c *Connection) smallSleep() {
	time.Sleep(time.Millisecond*100)
}

func (c  *Connection) write(b []byte) error {
	_, err := c.dev.Write(b)
	return err
}

func (c *Connection) writeBinaryFile(data []byte) error {
	_, err := c.dev.Write([]byte{0x02})
	if err != nil {
		return err
	}

	n, err := c.dev.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("only wrote %d rather than %d", n, len(data))
	}
	
	_, err = c.dev.Write([]byte{0x03, '\r'})
	if err != nil {
		return err
	}

	return nil
}

func checksum(data []byte) uint64 {
	u := uint64(1)
	for _ , n := range data {
		if (u & 0x80000000) != 0 {
			u = (u << 1) ^ 0x1d872b41
		} else {
			u = u << 1
		}
		u = (u ^ uint64(n)) & 0xFFFFFFFF
	}
	return u
}
