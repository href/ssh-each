package ssh

import (
	"fmt"
	"strconv"
)

// Destination describes an SSH target with a hostname (mandatory), a user
// (optional) and a port (optional, defaults to 22)
type Destination struct {
	Host string
	User string
	Port uint16
}

// String returns the host detination ([user@]host[:port]).
func (t Destination) String() string {
	switch {
	case t.Host == "":
		return ""
	case t.User == "" && t.Port == 0:
		return t.Host
	case t.User == "" && t.Port != 0:
		return fmt.Sprintf("%s:%d", t.Host, t.Port)
	case t.User != "" && t.Port == 0:
		return fmt.Sprintf("%s@%s", t.User, t.Host)
	default:
		return fmt.Sprintf("%s@%s:%d", t.User, t.Host, t.Port)
	}
}

// StringWithoutPort returns the host destination, but does not
// include a port, even if one is set
func (t Destination) StringWithoutPort() string {
	switch {
	case t.User == "":
		return t.Host
	default:
		return fmt.Sprintf("%s@%s", t.User, t.Host)
	}
}

// ParseDestination returns a destination from string, if not empty:
// - An empty string returns nil
// - An invalid port is ignored
func ParseDestination(text string) *Destination {
	if text == "" {
		return nil
	}

	// s -> position of @ in foo@bar
	// e -> position of : in bar:baz
	var s, e int = -1, -1
	for ix, char := range text {
		switch {
		case s > -1 && e > -1:
			break
		case s == -1 && char == '@':
			s = ix
		case e == -1 && char == ':':
			e = ix
		}
	}

	// without positions, all text is host
	if s == -1 && e == -1 {
		return &Destination{
			Host: text,
		}
	}

	// with no end position, there's user and host
	if e == -1 {
		return &Destination{
			User: text[:s],
			Host: text[s+1:],
		}
	}

	// with no start position, there's host and port
	if s == -1 {
		if port, _ := strconv.Atoi(text[e+1:]); port > 0 && port <= 65535 {
			return &Destination{
				Host: text[:e],
				Port: uint16(port),
			}
		}

		return &Destination{
			Host: text,
		}
	}

	// with both positions, there's user, host, and port
	if port, _ := strconv.Atoi(text[e+1:]); port > 0 && port <= 65535 {
		return &Destination{
			User: text[:s],
			Host: text[s+1 : e],
			Port: uint16(port),
		}
	}

	return &Destination{
		User: text[:s],
		Host: text[s+1 : e],
	}
}
