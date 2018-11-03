// Conway's Game of Life on an APC mini MIDI controller
package main

import (
	"./game"
	"./portmidi"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

type APCMini struct {
	midiIn  *portmidi.Stream
	midiOut *portmidi.Stream
}

// Assumes it's the only MIDI controller connected
func NewAPCMini() *APCMini {
	portmidi.Initialize()

	midiIn, err := portmidi.NewInputStream(portmidi.DefaultInputDeviceID(), 1024)
	if err != nil {
		log.Fatal(err)
	}

	midiOut, err := portmidi.NewOutputStream(portmidi.DefaultOutputDeviceID(), 1024, 0)
	if err != nil {
		log.Fatal(err)
	}

	return &APCMini{midiIn: midiIn, midiOut: midiOut}
}

func (a *APCMini) WriteShort(x, y, z int64) {
	a.midiOut.WriteShort(x, y, z)
}

func (a *APCMini) Clear() {
	for i := 0; i <= 99; i++ {
		a.midiOut.WriteShort(0x90, int64(i), 0)
	}
}

func (a *APCMini) Listen() <-chan []portmidi.Event {
	return a.midiIn.Listen()
}

func (a *APCMini) Close() {
	portmidi.Terminate()
}

func buttonIdFromXY(x int, y int) int64 {
	return int64(56 - y*8 + x)
}

func LifeToMidi(l *game.Life, apc *APCMini) {
	for y := 0; y < l.H(); y++ {
		for x := 0; x < l.W(); x++ {
			if l.Alive(x, y) {
				apc.WriteShort(0x90, buttonIdFromXY(x, y), 1)
			} else {
				apc.WriteShort(0x90, buttonIdFromXY(x, y), 0)
			}
		}
	}
}

func main() {
	apc := NewAPCMini()
	defer apc.Close()
	apc.Clear()

	life := game.NewLife(8, 8)

	go func() {
		for true {
			LifeToMidi(life, apc)
			fmt.Print("****************************************\n", life) // Clear screen and print field.
			life.Step()
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		keep_running := true
		midiin_chan := apc.Listen()
		for keep_running {
			events := <-midiin_chan
			fmt.Printf("events: %v\n", events)
			for _, event := range events {
				if keep_running {
					keep_running = event.Status != 128 || event.Data1 != 82 || event.Data2 != 127
				}

				if event.Status == 144 && event.Data1 == 89 && event.Data2 == 127 {
					life.Clear()
					LifeToMidi(life, apc)
				} else {
					if event.Status == 144 && event.Data2 == 127 {
						keyId := event.Data1
						x := int(keyId % 8)
						y := int(math.Abs(float64(keyId/8) - 7))

						life.Toggle(x, y)
					}
				}
			}

			LifeToMidi(life, apc)
		}
		os.Exit(0)
	}()

	select {}

	// for i := 0; i < 96; i++ {
	// 	out.WriteShort(0x90, int64(i), 0)
	// }
	// fmt.Printf("off\n")

	// // note on events to play C major chord
	// for i := 0; i <= 98; i++ {
	// 	out.WriteShort(0x90, int64(i-1), 01)
	// 	time.Sleep(100 * time.Millisecond)
	// 	fmt.Printf(".")
	// }

}
