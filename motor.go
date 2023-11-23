package viambuildhat

import (
	"context"
	"fmt"

	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func NewMotor(c *Connection, port int, logger logging.Logger) (motor.Motor, error) {
	m := &legoMotor{
		c:      c,
		logger: logger,
		port:   port,
	}

	return m, nil
}

type legoMotor struct {
	resource.AlwaysRebuild

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
	panic(7)
}

func (m *legoMotor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic(8)
}
