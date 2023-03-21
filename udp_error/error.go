package udp_error

type InvalidLength struct {
}

func (e InvalidLength) Error() string {
	return "InvalidLength"
}

type InvalidMagicNumber struct {
}

func (e InvalidMagicNumber) Error() string {
	return "InvalidMagicNumber"
}

type SessionNotMatch struct {
}

func (e SessionNotMatch) Error() string {
	return "SessionNotMatch"
}

type RecvNotComplete struct {
}

func (e RecvNotComplete) Error() string {
	return "RecvNotComplete"
}
