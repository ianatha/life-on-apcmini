package portmidi

// #cgo LDFLAGS: -lportmidi
//
// #include <stdlib.h>
// #include <portmidi.h>
// #include <porttime.h>
import "C"

import (
	"errors"
	"time"
	"unsafe"
)

const (
	minEventBufferSize = 1
	maxEventBufferSize = 1024
)

var (
	ErrMaxBuffer         = errors.New("portmidi: max event buffer size is 1024")
	ErrMinBuffer         = errors.New("portmidi: min event buffer size is 1")
	ErrInputUnavailable  = errors.New("portmidi: input is unavailable")
	ErrOutputUnavailable = errors.New("portmidi: output is unavailable")
)

type Channel int

type Event struct {
	Timestamp Timestamp
	Status    int64
	Data1     int64
	Data2     int64
}

type Stream struct {
	deviceID DeviceID
	pmStream *C.PmStream
}

func NewInputStream(id DeviceID, bufferSize int64) (stream *Stream, err error) {
	var str *C.PmStream
	errCode := C.Pm_OpenInput(
		(*unsafe.Pointer)(unsafe.Pointer(&str)),
		C.PmDeviceID(id), nil, C.int32_t(bufferSize), nil, nil)
	if errCode != 0 {
		return nil, convertToError(errCode)
	}
	if info := Info(id); !info.IsInputAvailable {
		return nil, ErrInputUnavailable
	}
	return &Stream{deviceID: id, pmStream: str}, nil
}

func NewOutputStream(id DeviceID, bufferSize int64, latency int64) (stream *Stream, err error) {
	var str *C.PmStream
	errCode := C.Pm_OpenOutput(
		(*unsafe.Pointer)(unsafe.Pointer(&str)),
		C.PmDeviceID(id), nil, C.int32_t(bufferSize), nil, nil, C.int32_t(latency))
	if errCode != 0 {
		return nil, convertToError(errCode)
	}
	if info := Info(id); !info.IsOutputAvailable {
		return nil, ErrOutputUnavailable
	}
	return &Stream{deviceID: id, pmStream: str}, nil
}

func (s *Stream) Close() error {
	if s.pmStream == nil {
		return nil
	}
	return convertToError(C.Pm_Close(unsafe.Pointer(s.pmStream)))
}

func (s *Stream) Abort() error {
	if s.pmStream == nil {
		return nil
	}
	return convertToError(C.Pm_Abort(unsafe.Pointer(s.pmStream)))
}

// Write writes a buffer of MIDI events to the output stream.
func (s *Stream) Write(events []Event) error {
	size := len(events)
	if size > maxEventBufferSize {
		return ErrMaxBuffer
	}
	buffer := make([]C.PmEvent, size)
	for i, evt := range events {
		var event C.PmEvent
		event.timestamp = C.PmTimestamp(evt.Timestamp)
		event.message = C.PmMessage((((evt.Data2 << 16) & 0xFF0000) | ((evt.Data1 << 8) & 0xFF00) | (evt.Status & 0xFF)))
		buffer[i] = event
	}
	return convertToError(C.Pm_Write(unsafe.Pointer(s.pmStream), &buffer[0], C.int32_t(size)))
}

// WriteShort writes a MIDI event of three bytes immediately to the output stream.
func (s *Stream) WriteShort(status int64, data1 int64, data2 int64) error {
	evt := Event{
		Timestamp: Timestamp(C.Pt_Time()),
		Status:    status,
		Data1:     data1,
		Data2:     data2,
	}
	return s.Write([]Event{evt})
}


// SetChannelMask filters incoming stream based on channel.
// In order to filter from more than a single channel, or multiple channels.
// s.SetChannelMask(Channel(1) | Channel(10)) will both filter input
// from channel 1 and 10.
func (s *Stream) SetChannelMask(mask int) error {
	return convertToError(C.Pm_SetChannelMask(unsafe.Pointer(s.pmStream), C.int(mask)))
}

// Reads from the input stream, the max number events to be read are
// determined by max.
func (s *Stream) read(max int) (events []Event, err error) {
	if max > maxEventBufferSize {
		return nil, ErrMaxBuffer
	}
	if max < minEventBufferSize {
		return nil, ErrMinBuffer
	}
	buffer := make([]C.PmEvent, max)
	numEvents := C.Pm_Read(unsafe.Pointer(s.pmStream), &buffer[0], C.int32_t(max))
	if numEvents < 0 {
		return nil, convertToError(C.PmError(numEvents))
	}
	events = make([]Event, numEvents)
	for i := 0; i < int(numEvents); i++ {
		events[i] = Event{
			Timestamp: Timestamp(buffer[i].timestamp),
			Status:    int64(buffer[i].message) & 0xFF,
			Data1:     (int64(buffer[i].message) >> 8) & 0xFF,
			Data2:     (int64(buffer[i].message) >> 16) & 0xFF,
		}
	}
	return
}

func (s *Stream) Listen() <-chan []Event {
	ch := make(chan []Event, 32)
	go func(s *Stream, ch chan []Event) {
		for {
			// sleep for a while before the new polling tick,
			// otherwise operation is too intensive and blocking
			time.Sleep(10 * time.Millisecond)
			events, err := s.read(1024)
			// Note: It's not very reasonable to push sliced data into
			// a channel, several perf penalities there are.
			// This function is added as a handy utility.
			if err != nil {
				continue
			}
			if (len(events) > 0) {
				ch <- events
			}
		}
	}(s, ch)
	return ch
}

func (s *Stream) Poll() (bool, error) {
	poll := C.Pm_Poll(unsafe.Pointer(s.pmStream))
	if poll < 0 {
		return false, convertToError(C.PmError(poll))
	}
	return poll > 0, nil
}