ssh-each
--------

Utility to run an SSH command against multiple servers concurrently.

## Examples

```bash
# Tail the journal on multiple servers 
$ ssh-each -servers frontend,backend --tty 'journalctl -f'
frontend: Aug 17 20:40:43 frontend nginx[1]: POST /login
backend: Aug 17 20:40:44 backend authsvc[2]: Logged in user <example>
...

# Read hostnames from stdin (another process, a file)
$ echo -e "web\ndatabase" | ssh-each --mode check 'systemctl is-enabled nginx' 
web: ✓
database: x

# Limit number of connections
$ cat many-servers.txt | ssh-each 'ping -c 1 8.8.8.8 | grep transmitted' --workers=5 --mode plain
1 packets transmitted, 1 received, 0% packet loss, time 0ms
1 packets transmitted, 1 received, 0% packet loss, time 0ms
...
```

## Usage

```bash
Usage: ssh-each [-t] [-s=<comma-separated-servers>] [-w=<workers>] [-u=<user>] [-p=<port>] [-m=<mode>] COMMAND

Run SSH commands on multiple servers concurrently.
Servers can be passed via -s/--servers, or STDIN.

Output Modes (-m/--mode):
  host: shows server before each outputted line, default
  plain: show output as-is
  check: show server and ✓ on success, x on failure, no output
  exit: show server and exit code, no output
  slient: show nothing

Exit Code:
  ssh-each will return an exit code of 0, if at least one command
  completed and all completed commands were successful.

Arguments:
  COMMAND         Command to execute

Options:
  -s, --servers   Comma separated servers
  -w, --workers   Concurrent SSH processes (default 16)
  -p, --port      Default port (default 0)
  -m, --mode      Output mode (default "host")
  -t, --tty       Use pseudo-terminal
  -u, --user      Default user
```

## Install
```bash
go install github.com/href/ssh-each@latest
```
