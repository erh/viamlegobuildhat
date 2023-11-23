package viambuildhat

import (
	"context"
	"fmt"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"
)

var MotorModel = family.WithModel("motor")

func init() {
	resource.RegisterComponent(
		motor.API,
		MotorModel,
		resource.Registration[motor.Motor, *MotorConfig]{
			Constructor: newMotor,
		})
}

type MotorConfig struct {
	Connection string
	Port       string
}

func (c *MotorConfig) Validate(path string) ([]string, error) {
	if c.Connection == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "connection")
	}

	_, err := c.portNumber(path)
	if err != nil {
		return nil, err
	}

	return []string{c.Connection}, nil
}

func (c *MotorConfig) portNumber(path string) (int, error) {
	if len(c.Port) != 1 {
		return 0, utils.NewConfigValidationError(path, fmt.Errorf("port has to be exactly 1 character long"))
	}
	x := c.Port[0]

	if x >= 'A' && x <= 'D' {
		return int(x - 'A'), nil
	}

	if x >= 'a' && x <= 'd' {
		return int(x - 'a'), nil
	}

	if x >= '0' && x <= '3' {
		return int(x - '0'), nil
	}

	return 0, utils.NewConfigValidationError(path, fmt.Errorf("invalid port [%s]", c.Port))
}

func newMotor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (motor.Motor, error) {
	newConf, err := resource.NativeConfig[*MotorConfig](conf)
	if err != nil {
		return nil, err
	}

	r, err := deps.Lookup(generic.Named(newConf.Connection))
	if err != nil {
		return nil, err
	}

	c, ok := r.(*Connection)
	if !ok {
		return nil, fmt.Errorf("connection isn't a connection, is a %T", r)
	}

	port, err := newConf.portNumber("")
	if err != nil {
		return nil, err
	}

	return NewMotor(conf.ResourceName(), c, port, logger)
}

func NewMotor(name resource.Name, c *Connection, port int, logger logging.Logger) (motor.Motor, error) {
	m := &legoMotor{
		name:   name,
		c:      c,
		logger: logger,
		port:   port,
	}

	return m, nil
}

type legoMotor struct {
	resource.AlwaysRebuild

	name   resource.Name
	c      *Connection
	logger logging.Logger
	port   int
}

func (m *legoMotor) SetPower(ctx context.Context, powerPct float64, extra map[string]interface{}) error {
	cmd := fmt.Sprintf("port %d; plimit 1; select 0; pwm; set %0.3f\r", m.port, powerPct)
	m.logger.Infof(cmd)
	return m.c.write([]byte(cmd))
}

func (m *legoMotor) GoFor(ctx context.Context, rpm, revolutions float64, extra map[string]interface{}) error {
	panic(6)
}

func (m *legoMotor) GoTo(ctx context.Context, rpm, positionRevolutions float64, extra map[string]interface{}) error {
	panic(5)
}

func (m *legoMotor) ResetZeroPosition(ctx context.Context, offset float64, extra map[string]interface{}) error {
	panic(4)
}

func (m *legoMotor) Position(ctx context.Context, extra map[string]interface{}) (float64, error) {
	panic(3)
}

func (m *legoMotor) Properties(ctx context.Context, extra map[string]interface{}) (motor.Properties, error) {
	panic(2)
}

func (m *legoMotor) IsPowered(ctx context.Context, extra map[string]interface{}) (bool, float64, error) {
	panic(1)
}

func (m *legoMotor) IsMoving(ctx context.Context) (bool, error) {
	powered, _, err := m.IsPowered(ctx, nil)
	return powered, err
}

func (m *legoMotor) Stop(ctx context.Context, extra map[string]interface{}) error {
	return m.SetPower(ctx, 0, extra)
}

func (m *legoMotor) Close(ctx context.Context) error {
	return m.SetPower(ctx, 0, nil)
}

func (m *legoMotor) Name() resource.Name {
	return m.name
}

func (m *legoMotor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic(8)
}
