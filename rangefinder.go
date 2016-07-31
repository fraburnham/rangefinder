package rangefinder

type rangefinder interface {
	Distance_cm() (float32, error) // return distance in cm
}
