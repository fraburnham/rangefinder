package rangefinder

import (
	"fmt"
	"gpio"
	"time"
)

const (
	triggerSleep = time.Duration(10 * time.Microsecond)
	// speed of sound could be improved with
	// a temp/humidity sensor and some gooder math
	// http://hyperphysics.phy-astr.gsu.edu/hbase/sound/souspe.html
	speed_of_sound_cm_s = 34300
)

type HCSR04 struct {
	triggerGPIO gpio.RpiGPIO
	signalGPIO gpio.RpiGPIO
}

func sendTrigger(triggerGPIO gpio.RpiGPIO) error {
	// handle errs
	triggerGPIO.WriteValue(1)
	time.Sleep(triggerSleep)
	triggerGPIO.WriteValue(0)
	return nil
}

func captureResponse(signalGPIO gpio.RpiGPIO) (time.Time, time.Time, error) {
	startTime := time.Now() // just in case? Maybe a bad idea
	var endTime time.Time

	// factor out the loop, also add timeouts or some other sanity check
	// maybe we missed the boat on the 0 ...
	for val, err := signalGPIO.ReadValue(); val == 0; val, err = signalGPIO.ReadValue() {
		startTime = time.Now()
		if err != nil {
			return time.Now(), time.Now(), err
		}
	}

	for val, err := signalGPIO.ReadValue(); val == 1; val, err = signalGPIO.ReadValue() {
		endTime = time.Now()
		if err != nil {
			return time.Now(), time.Now(), err
		}
	}

	return startTime, endTime, nil // return some Timeout error
}

func calculateDistace(startTime time.Time, endTime time.Time) float32 {
	durationBetween := endTime.Sub(startTime)
	fmt.Println("Time between millis: ", durationBetween) // sometimes this goes negative...
	return float32((durationBetween.Seconds() * speed_of_sound_cm_s) / 2)
}

func (h *HCSR04) Distance_cm() (float32, error) {
	// error handling
	sendTrigger(h.triggerGPIO)
	startTime, endTime, _ := captureResponse(h.signalGPIO)
	return calculateDistace(startTime, endTime), nil
}

func NewHCSRO4(triggerPin int, signalPin int) (*HCSR04, error) {
	triggerGPIO, err := gpio.NewRpiOutput(triggerPin)
	if err != nil {
		return nil, err
	}

	signalGPIO, err := gpio.NewRpiInput(signalPin)
	if err != nil {
		return nil, err
	}

	return &HCSR04{triggerGPIO: *triggerGPIO,
		signalGPIO: *signalGPIO}, nil
}
