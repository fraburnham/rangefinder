package rangefinder

import (
	"fmt"
	"gpio"
	"gpio/events"
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
	triggerGPIO gpio.GPIO
	signalGPIO gpio.GPIO
}

func sendTrigger(triggerGPIO gpio.GPIO) error {
	// handle errs
	triggerGPIO.WriteValue(1)
	time.Sleep(triggerSleep)
	triggerGPIO.WriteValue(0)
	return nil
}

func captureResponse(signalGPIO gpio.GPIO) (time.Time, time.Time, error) {
	eventCh, ctrlCh := events.StartEdgeTrigger(signalGPIO)

	startEvent := <-eventCh
	// check the event
	endEvent := <-eventCh

	events.StopEdgeTrigger(ctrlCh)
	return startEvent.Timestamp, endEvent.Timestamp, nil // return some Timeout error
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

	return &HCSR04{triggerGPIO: triggerGPIO,
		signalGPIO: signalGPIO}, nil
}
