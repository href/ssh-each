package ssh

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// CommandBuilder helps build exec.Cmd instances using a single command for
// many servers.
type CommandBuilder struct {
	// TTY is true if a pseudo terminal should be attached.
	TTY bool

	// User to use for all connections (unless set on the destination). If
	// left empty, and no user is provided on the destination, no user is
	// passed, leaving the SSH command to chose it.
	ExplicitUser string

	// Port to use for all connections (unless set on the destination). If
	// left at 0, and no port is provided on the destination, no port is
	// passed, leaving the SSH command to chose it.
	ExplicitPort uint16

	// Command to be executed each time.
	Command string
}

// LinkedCommand is a command linked to a server
type LinkedCommand struct {
	Command *exec.Cmd
	Server  string
}

// For creates a command for the given destination
func (o *CommandBuilder) For(ctx context.Context, dst Destination) *exec.Cmd {
	// We'll heave at least "ssh", a host, and a command.
	args := make([]string, 0, 3)
	args = append(args, "ssh")

	switch {
	case dst.Port > 0:
		args = append(args, "-p", strconv.Itoa(int(dst.Port)))
	case o.ExplicitPort > 0:
		args = append(args, "-p", strconv.Itoa(int(o.ExplicitPort)))
	}

	// If a TTY is wanted, use the '-tt' variant, as the weaker '-t' variant
	// won't work since we are not forwarding STDIN
	if o.TTY {
		args = append(args, "-tt")
	}

	// If the destination has a user, we don't need to set it anywhere, as
	// it will be rendered as "user@host", which has precedence over everything
	// else.
	if dst.User == "" && o.ExplicitUser != "" {
		args = append(args, "-l", o.ExplicitUser)
	}

	// Finally, we add the host and the command
	args = append(args, dst.StringWithoutPort(), o.Command)

	return exec.CommandContext(ctx, args[0], args[1:]...)
}

// FromReader builds commands for the servers read from the reader. Each
// line is expected to include a single server.
func (o *CommandBuilder) FromReader(
	ctx context.Context,
	r io.Reader,
) <-chan LinkedCommand {
	ch := make(chan LinkedCommand)

	go func() {
		defer close(ch)

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			server := scanner.Text()
			server = strings.Trim(server, " \n")

			if server == "" {
				continue
			}

			linked := LinkedCommand{
				Command: o.For(ctx, *ParseDestination(server)),
				Server:  server,
			}

			select {
			case ch <- linked:
				continue
			case <-ctx.Done():
				break
			}
		}
	}()

	return ch
}
