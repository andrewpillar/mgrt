package revision

import "errors"

const (
	Up Direction	= iota
	Down
)

type Direction int64

func (d Direction) String() string {
	switch d {
		case Up:
			return "up"
		case Down:
			return "down"
		default:
			return "unknown"
	}
}

func (d *Direction) Scan(src interface{}) error {
	i, ok := src.(int64)

	if !ok {
		return errors.New("failed to scan direction type: could not type assert to int64")
	}

	(*d) = Direction(i)

	return nil
}

func (d Direction) Invert() Direction {
	if d == Up {
		return Down
	}

	return Up
}
