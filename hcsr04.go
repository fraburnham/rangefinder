package rangefinder

import (
	"github.com/fraburnham/gpio"
	"time"
)

const (
	triggerSleep = time.Duration(10 * time.Microsecond)
	// speed of sound could be improved with
	// a temp/humidity sensor and some gooder math
	// http://hyperphysics.phy-astr.gsu.edu/hbase/sound/souspe.html
	speedOfSoundCmPerSec = 34300
)

type HCSR04 struct {
	trigger         gpio.GPIO
	signal          gpio.GPIO
	currentDistance float32
	ctrlCh          chan bool
	err             error // for now, it would be nice to update this to a custom error
}

func waitForInputLow(pin gpio.GPIO) error {
	val, err := pin.ReadValue()
	for val != 0 {
		val, err = pin.ReadValue()
		if err != nil {
			return err
		}
	}
	return err
}

func sendTrigger(trigger gpio.GPIO) error {
	err := trigger.WriteValue(1)
	if err != nil {
		return err
	}

	time.Sleep(triggerSleep)

	err = trigger.WriteValue(0)
	if err != nil {
		return err
	}

	return nil
}

func calculateDistace(startTime time.Time, endTime time.Time) float32 {
	durationBetween := endTime.Sub(startTime)
	return float32((durationBetween.Seconds() * speedOfSoundCmPerSec) / 2)
}

func distanceUpdater(h *HCSR04) {
	err := waitForInputLow(h.signal)
	if err != nil {
		h.err = err
	}

	interruptCh := make(chan gpio.InterruptEvent)

	err = h.signal.SetInterrupt("both", interruptCh, 50)
	if err != nil {
		h.err = err
	}

	for {
		var highInterrupt gpio.InterruptEvent
		var done bool

		err = sendTrigger(h.trigger)
		if err != nil {
			h.err = err
		}

		for !done {
			select {
			case <-h.ctrlCh:
				h.signal.ClearInterrupt()
				h.currentDistance = 0
				return
			case interrupt := <-interruptCh:
				switch interrupt.Value {
				case 1:
					highInterrupt = interrupt
				case 0:
					h.currentDistance = calculateDistace(highInterrupt.Timestamp, interrupt.Timestamp)
					done = true
				}

			}
		}

		// polling duration should be configurable
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
}

func (h *HCSR04) DistanceCm() (float32, error) {
	err := h.err
	// clear the error for the next call so clients don't get stuck in error loops
	h.err = nil
	return h.currentDistance, err
}

func (h *HCSR04) Close() {
	h.ctrlCh <- true // should probably add a WaitGroup here to wait for the routine to cleanup
}

// need to correct package layout like gpio, separate implementation from interface like god intended
func New(trigger gpio.GPIO, signal gpio.GPIO) (*HCSR04, error) {
	// need to get the polling frequency from the user

	h := &HCSR04{
		trigger: trigger,
		signal:  signal,
		ctrlCh: make(chan bool)}

	go distanceUpdater(h)

	return h, nil
}
