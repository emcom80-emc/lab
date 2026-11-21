package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// RunREPL reads lines from r and dispatches slash-commands to session.
// Output is written to w. Returns when /exit is issued or r reaches EOF.
func RunREPL(session *Session, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "/") {
			fmt.Fprintln(w, "hint: commands start with /  (try /help)")
			continue
		}
		out, err := session.Command(line)
		if err != nil {
			fmt.Fprintf(w, "error: %v\n", err)
			continue
		}
		if out == ExitSentinel {
			fmt.Fprintln(w, "Goodbye!")
			return nil
		}
		if out != "" {
			fmt.Fprintln(w, out)
		}
	}
	return scanner.Err()
}
