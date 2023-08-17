package ssh

import (
	"context"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestCommandBuilder(t *testing.T) {
	assert := func(cb CommandBuilder, dst Destination, expected []string) {
		cmd := cb.For(context.Background(), dst)

		if diff := deep.Equal(cmd.Args, expected); diff != nil {
			t.Error(diff)
		}
	}

	assert(
		CommandBuilder{Command: "command"},
		Destination{Host: "host"},
		[]string{"ssh", "host", "command"},
	)

	assert(
		CommandBuilder{Command: "command", ExplicitUser: "user"},
		Destination{Host: "host"},
		[]string{"ssh", "-l", "user", "host", "command"},
	)

	assert(
		CommandBuilder{Command: "command", ExplicitPort: 1234},
		Destination{Host: "host"},
		[]string{"ssh", "-p", "1234", "host", "command"},
	)

	assert(
		CommandBuilder{Command: "command", ExplicitUser: "user"},
		Destination{Host: "host", User: "host-user"},
		[]string{"ssh", "host-user@host", "command"},
	)

	assert(
		CommandBuilder{Command: "command", ExplicitPort: 1234},
		Destination{Host: "host", Port: 4567},
		[]string{"ssh", "-p", "4567", "host", "command"},
	)
}

func TestCommandBuilderFromReader(t *testing.T) {
	reader := strings.NewReader("host1\nhost2")

	cb := CommandBuilder{Command: "whoami"}
	ch := cb.FromReader(context.Background(), reader)

	expected := [][]string{
		{"ssh", "host1", "whoami"},
		{"ssh", "host2", "whoami"},
	}

	produced := [][]string{}
	for cmd := range ch {
		produced = append(produced, cmd.Command.Args)
	}

	if diff := deep.Equal(expected, produced); diff != nil {
		t.Error(diff)
	}
}
