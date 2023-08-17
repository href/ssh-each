package ssh

import (
	"github.com/go-test/deep"
	"testing"
)

func TestParseDestination(t *testing.T) {
	assert := func(input string, expected Destination) {
		parsed := ParseDestination(input)

		if diff := deep.Equal(*parsed, expected); diff != nil {
			t.Error(diff)
		}
	}

	assert("foo", Destination{Host: "foo"})
	assert("foo@bar", Destination{User: "foo", Host: "bar"})
	assert("foo@bar@baz", Destination{User: "foo", Host: "bar@baz"})
	assert("foo@bar:123", Destination{User: "foo", Host: "bar", Port: 123})
	assert("bar:123", Destination{Host: "bar", Port: 123})
	assert("bar:123123123", Destination{Host: "bar:123123123"})
	assert("foo:bar", Destination{Host: "foo:bar"})
}

func TestFormatDestination(t *testing.T) {
	assert := func(input Destination, expected string) {
		formatted := input.String()

		if formatted != expected {
			t.Errorf("%v: %s format, expected %s", input, formatted, expected)
		}
	}

	assert(Destination{Host: "foo"}, "foo")
	assert(Destination{User: "foo", Host: "bar"}, "foo@bar")
	assert(Destination{User: "foo", Host: "bar", Port: 22}, "foo@bar:22")
}
