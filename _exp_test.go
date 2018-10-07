package snd

import (
	"fmt"
	"math"
	"testing"
)

func TestRadian(t *testing.T) {

	var rad Radian
	const n = 10
	onethird := Radian(math.MaxUint64 / 3.)
	for i := 1; i < n; i++ {
		rad += onethird + onethird + onethird

		fmt.Println(rad.Degrees())
		fmt.Println(math.MaxUint64 - rad)
		fmt.Println()

		// if rad.Degrees() != 360 {
		// t.Fatalf("failed on iteration %v", i)
		// }
	}

	// var rad Radian2

	// fmt.Println(RPiTwo % 3)
	// x := float64(RPiTwo) / 3.0

	// onethird := Radian2(x)
	// fmt.Println(onethird)
	// for i := 0; i < 10; i++ {
	// rad += onethird
	// fmt.Println(rad.Degrees())
	// }

	// var z Radian2

	// const n = 44100
	// for i := 1; i < n; i++ {
	// z += RPiTwo
	// if uint64(z)%uint64(RPiTwo) != 0 {
	// t.Fatalf("%v: mod failed for %v", i, z)
	// }
	// if uint64(i) != z.Hertz() {
	// t.Fatalf("%v: hertz failed for %v", i, z)
	// }
	// t.Log(z.Degrees(), z.Hertz(), uint64(z)%uint64(RPiTwo))
	// }

	// z += 1
	// t.Log(z.Degrees(), z.Hertz(), uint64(z)%uint64(RPiTwo))

	// x := uint64(math.MaxUint64)
	// for ; x%uint64(RPiTwo) != 0; x-- {
	// }
	// t.Log(x % uint64(RPiTwo))
	// t.Logf("0x%016X", x)

	// THz GHz MHz kHz  Hz
	// 281,474,976,710,655
	// t.Log(Radian2(MaxHertz).Hertz())

	// 281,474,976,645,120
	// t.Log(Radian2(MaxHertz2).Hertz())
	// 4,294,967,295
	// t.Log(MaxHertz2 / 0x100000000)
	// t.Log(MaxHertz3 / 0x1000000000000)

	//           mHz μHz nHz pHz fHz
	// 000000000.000,000,000,232,830,64365386963e-10
	// t.Log(1. / (1. + math.MaxUint32))

	// t.Log(RPiTwo)

	// t.Log(2. / 65536)

	// 0.000,015,258,789,0625

	// t.Log(RPi.Degrees())
	// t.Log(math.MaxUint64 / RPi)

	// x := 2 * RPi
	// t.Logf("0x%016X", x)

	// 562,949,953,421,311

	// a := 750 * Millihertz
	// r := a.Angular()
	// t.Log(r.Degrees() / 360.0 / float64(HHertz))

	// r2 := Radian(a)
	// t.Log(r2)
	// t.Log(r2.Degrees())
}
