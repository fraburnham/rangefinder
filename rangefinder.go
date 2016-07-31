// Provide a simple interface for implementing rangefinders.
package rangefinder

type Rangefinder interface {
	DistanceCm() (float32, error)
}
