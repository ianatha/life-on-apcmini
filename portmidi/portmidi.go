package portmidi

// #cgo CFLAGS:  -I/usr/local/include
// #cgo LDFLAGS: -lportmidi -L/usr/local/lib
//
// #include <stdlib.h>
// #include <portmidi.h>
// #include <porttime.h>
import "C"

import (
	"errors"
)

// DeviceID is a MIDI device ID.
type DeviceID int

// DeviceInfo provides info about a MIDI device.
type DeviceInfo struct {
	Interface         string
	Name              string
	IsInputAvailable  bool
	IsOutputAvailable bool
	IsOpened          bool
}

type Timestamp int64

func Initialize() error {
	if code := C.Pm_Initialize(); code != 0 {
		return convertToError(code)
	}
	C.Pt_Start(C.int(1), nil, nil)
	return nil
}

func Terminate() error {
	C.Pt_Stop()
	return convertToError(C.Pm_Terminate())
}

func DefaultInputDeviceID() DeviceID {
	return DeviceID(C.Pm_GetDefaultInputDeviceID())
}

func DefaultOutputDeviceID() DeviceID {
	return DeviceID(C.Pm_GetDefaultOutputDeviceID())
}

func CountDevices() int {
	return int(C.Pm_CountDevices())
}

func Info(deviceID DeviceID) *DeviceInfo {
	info := C.Pm_GetDeviceInfo(C.PmDeviceID(deviceID))
	if info == nil {
		return nil
	}
	return &DeviceInfo{
		Interface:         C.GoString(info.interf),
		Name:              C.GoString(info.name),
		IsInputAvailable:  info.input > 0,
		IsOutputAvailable: info.output > 0,
		IsOpened:          info.opened > 0,
	}
}

func Time() Timestamp {
	return Timestamp(C.Pt_Time())
}

func convertToError(code C.PmError) error {
	if code >= 0 {
		return nil
	}
	return errors.New(C.GoString(C.Pm_GetErrorText(code)))
}