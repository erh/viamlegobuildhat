package viambuildhat

import (
	"testing"

	"go.viam.com/test"
)

func TestChecksum(t *testing.T) {
	x := []byte{0xE, 0x5, 0x1, 0x7, 0x31, 0xE, 0x5, 0x1, 0x7, 0x31, 0xE, 0x5, 0x1, 0x7, 0x31}
	test.That(t, checksum(x), test.ShouldEqual, 217747)

	test.That(t, len(firmware), test.ShouldEqual, 54168)
	test.That(t, checksum(firmware), test.ShouldEqual, 2080457803)

	
}
