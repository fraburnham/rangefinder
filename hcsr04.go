package rangefinder

import (
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

type eventLength struct {
	startTime time.Time
	endTime time.Time
	err error
}

func waitForInputLow(pin gpio.GPIO) error {
	val, err := pin.ReadValue()
	for val != 1 {
		val, err = pin.ReadValue()
		if err != nil {
			return err
		}
	}
	return err
}

func sendTrigger(triggerGPIO gpio.GPIO) error {
	// handle errs
	triggerGPIO.WriteValue(1)
	time.Sleep(triggerSleep)
	triggerGPIO.WriteValue(0)
	return nil
}

func captureResponse(signalGPIO gpio.GPIO, resultCh chan eventLength) {
	eventCh, ctrlCh := events.StartEdgeTrigger(signalGPIO, 2)

	//ready message
	resultCh <- eventLength{}

	startEvent := <-eventCh
	// check the event
	endEvent := <-eventCh

	events.StopEdgeTrigger(ctrlCh)
	resultCh <- eventLength{startTime: startEvent.Timestamp,
		endTime: endEvent.Timestamp,
		err: nil} // return some Timeout error
}

func calculateDistace(startTime time.Time, endTime time.Time) float32 {
	durationBetween := endTime.Sub(startTime)
	return float32((durationBetween.Seconds() * speed_of_sound_cm_s) / 2)
}

func (h *HCSR04) Distance_cm() (float32, error) {
	// error handling
	waitForInputLow(h.signalGPIO)
	resultCh := make(chan eventLength)
	go captureResponse(h.signalGPIO, resultCh)
	<-resultCh // make sure stuff is started
	sendTrigger(h.triggerGPIO)
	length := <-resultCh
	// error handling!
	return calculateDistace(length.startTime, length.endTime), nil
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
