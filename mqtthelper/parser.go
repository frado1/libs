package mqtthelper

func ParseSetSystemState(b []byte) (SetSystemState, error) {
	s := SetSystemState(string(b))
	if err := s.Validate(); err != nil {
		return SetSystemState(""), err
	}

	return s, nil
}

func ParseSetState(b []byte) (SetState, error) {
	s := SetState(string(b))
	if err := s.Validate(); err != nil {
		return SetState(""), err
	}

	return s, nil
}
