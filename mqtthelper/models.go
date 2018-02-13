package mqtthelper

import (
	"fmt"
)

type SetSystemState string

func (s SetSystemState) Validate() error {
	switch s {
	case "connect":
		return nil
	case "disconnect":
		return nil
	default:
		return fmt.Errorf("System state '%s' is not valid", s)
	}
}

type SetState string

func (s SetState) Validate() error {
	switch s {
	case "on":
		return nil
	case "off":
		return nil
	default:
		return fmt.Errorf("State '%s' is not valid", s)
	}
}
