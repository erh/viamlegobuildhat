package viambuildhat

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"go.uber.org/multierr"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var ConnectionModel = family.WithModel("connection")

func init() {
	resource.RegisterComponent(
		generic.API,
		ConnectionModel,
		resource.Registration[resource.Resource, resource.NoNativeConfig]{
			Constructor: newConnection,
		})
}

//go:embed data/firmware.bin
var firmware []byte

//go:embed data/signature.bin
var signature []byte

const defaultPath = "/dev/serial0"

const lineCacheSize = 5

func newConnection(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (resource.Resource, error) {
	path := defaultPath
	if config.Attributes.Has("path") {
		path = config.Attributes.String("path")
	}
	return GetConnection(ctx, config.ResourceName(), path, logger)
}

type globalConnCache struct {
	mu    sync.Mutex
	conns map[string]*Connection
}

var theGlobalConnCache globalConnCache

// path is usually /dev/serial0
func GetConnection(ctx context.Context, name resource.Name, path string, logger logging.Logger) (*Connection, error) {
	if path == "" {
		path = defaultPath
	}

	theGlobalConnCache.mu.Lock()
	defer theGlobalConnCache.mu.Unlock()
	if theGlobalConnCache.conns == nil {
		theGlobalConnCache.conns = map[string]*Connection{}
	}
	conn := theGlobalConnCache.conns[path]
	if conn != nil {
		return conn, nil
	}

	options := serial.OpenOptions{
		PortName:          path,
		BaudRate:          115200,
		DataBits:          8,
		StopBits:          1,
		MinimumReadSize:   1,
		RTSCTSFlowControl: true,
	}

	dev, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	conn = &Connection{
		name:       name,
		logger:     logger,
		dev:        dev,
		portStatus: map[int]string{},
	}
	conn.closed.Store(false)
	conn.smallSleep()

	go conn.readLoop()
	conn.smallSleep()

	err = conn.init(false)
	if err != nil {
		return nil, multierr.Combine(conn.Close(ctx), err)
	}

	theGlobalConnCache.conns[path] = conn

	return conn, nil
}

type Connection struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	closed atomic.Bool

	devLock sync.Mutex
	dev     io.ReadWriteCloser

	metaDataLock  sync.Mutex
	versionString string
	lastLines     []string
	portStatus    map[int]string
}

func (c *Connection) init(secondTime bool) error {
	err := c.write([]byte("version\r"))
	if err != nil {
		return err
	}

	time.Sleep(20 * time.Millisecond)

	for i := 0; i < 30 && c.version() == ""; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	v := c.version()
	if v == "" {
		return fmt.Errorf("no version string detected")
	}

	if strings.Contains(v, "bootloader") {
		c.logger.Infof("loading firmware because bootloader: %s", v)
		if secondTime {
			return fmt.Errorf("still got bootloader on second time")
		}
		err = c.loadFirmware()
		if err != nil {
			return err
		}
		return c.init(true)
	}

	err = c.write([]byte("echo 0\r"))
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) version() string {
	c.metaDataLock.Lock()
	defer c.metaDataLock.Unlock()
	return c.versionString
}

func (c *Connection) readLoop() {
	buf := bufio.NewReader(c.dev)
	for !c.closed.Load() {
		line, err := buf.ReadString('\r')
		if err != nil {
			c.logger.Errorf("got error in readLoop %v: ", err)
			if err == io.ErrClosedPipe {
				panic("TODO - ErrClosedPipe - should i reconnect")
			}
			panic(err)
		}

		c.handleLogLine(line)

	}
}

func (c *Connection) handleLogLine(line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	if strings.HasPrefix(line, "BHBL>") {
		// prompt or echo
		return
	}

	if strings.Contains(line, " version") {
		c.metaDataLock.Lock()
		c.versionString = line
		c.metaDataLock.Unlock()
		c.logger.Infof("firmware version at startup: %s", line)
		return
	}

	if line[0] == 'P' {
		if len(line) > 3 {
			portid := int(line[1] - '0')
			status := line[2:]

			c.metaDataLock.Lock()
			defer c.metaDataLock.Unlock()

			old := c.portStatus[portid]
			c.portStatus[portid] = status
			if old != status {
				c.logger.Infof("port status change %s", line)
			}
			return
		}
		c.logger.Infof("confusing port status line [%s]", line)
		return
	}

	c.logger.Infof("got unknown line [%s]", line)

	c.metaDataLock.Lock()
	defer c.metaDataLock.Unlock()
	c.lastLines = append(c.lastLines, line)
	if len(c.lastLines) > lineCacheSize {
		c.lastLines = c.lastLines[len(c.lastLines)-lineCacheSize:]
	}
}

func (c *Connection) waitForLine(line string) error {
	for i := 0; i < 100; i++ {

		c.metaDataLock.Lock()
		for _, l := range c.lastLines {
			if strings.Contains(l, line) {
				c.metaDataLock.Unlock()
				return nil
			}
		}

		c.metaDataLock.Unlock()
		time.Sleep(time.Millisecond * 100)
	}
	return fmt.Errorf("did not get line [%s]", line)
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

	err = c.write([]byte("reboot\r"))
	if err != nil {
		return err
	}

	err = c.waitForLine("Done initialising ports")
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second) // TODO - this works but is ugly, cribbed from python system

	return nil
}

func (c *Connection) Close(context.Context) error {
	c.closed.Store(true)

	// TODO:

	turnoff := ""
	for p := 0; p < 4; p++ {
		// if p type == 64
		// hexstr = ' '.join(f'{h:x}' for h in [0xc2, 0, 0, 0, 0, 0, 0, 0, 0, 0])
		// self.write(f"port {p} ; write1 {hexstr}\r".encode())

		turnoff = fmt.Sprintf("%s port %d ; pwm off ; coast ; off ;", p)
	}
	c.write([]byte(turnoff + "\r"))
	c.write([]byte("port 0 ; select ; port 1 ; select ; port 2 ; select ; port 3 ; select ; echo 0\r"))
	return c.dev.Close()
}

func (c *Connection) smallSleep() {
	time.Sleep(time.Millisecond * 100)
}

func (c *Connection) write(b []byte) error {
	c.devLock.Lock()
	defer c.devLock.Unlock()
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

func (c *Connection) Name() resource.Name {
	return c.name
}

func (c *Connection) DoCommand(ctx context.Context, cmdMessage map[string]interface{}) (map[string]interface{}, error) {
	cmd, has := cmdMessage["cmd"]
	if has {
		err := c.write([]byte(strings.TrimSpace(fmt.Sprintf("%v", cmd)) + "\r"))
		if err != nil {
			return nil, err
		}
		// TODO: capture output
		return map[string]interface{}{}, nil
	}
	_, has = cmdMessage["lastLines"]
	if has {
		c.metaDataLock.Lock()
		defer c.metaDataLock.Unlock()
		s := ""
		for _, l := range c.lastLines {
			s += l + "\n"
		}
		return map[string]interface{}{"lastLines": s}, nil
	}
	return nil, fmt.Errorf("viambuildhat connection can't handle command %v", cmd)
}

func checksum(data []byte) uint64 {
	u := uint64(1)
	for _, n := range data {
		if (u & 0x80000000) != 0 {
			u = (u << 1) ^ 0x1d872b41
		} else {
			u = u << 1
		}
		u = (u ^ uint64(n)) & 0xFFFFFFFF
	}
	return u
}
