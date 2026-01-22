package tool

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Bash struct {
	sessions map[string]*BashSession
	mu       sync.RWMutex
}

type BashSession struct {
	process   *exec.Cmd
	started   bool
	timedOut  bool
	command   string
	outputDelay time.Duration
	timeout   time.Duration
	sentinel  string
	stdin     *bufio.Writer
	stdout    *bufio.Reader
	stderr    *bufio.Reader
}

func NewBash() *Bash {
	return &Bash{
		sessions: make(map[string]*BashSession),
	}
}

func (b *Bash) Name() string {
	return "bash"
}

func (b *Bash) Description() string {
	return `Execute a bash command in the terminal.
* Long running commands: For commands that may run indefinitely, it should be run in the background and the output should be redirected to a file, e.g. command = "python3 app.py > server.log 2>&1 &".
* Interactive: If a bash command returns exit code -1, this means the process is not yet finished. The assistant must then send a second call to terminal with an empty "command" (which will retrieve any additional logs), or it can send additional text (set "command" to the text) to STDIN of the running process, or it can send command="ctrl+c" to interrupt the process.
* Timeout: If a command execution result says "Command timed out. Sending SIGINT to the process", the assistant should retry running the command in the background.`
}

func (b *Bash) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The bash command to execute. Use empty string to retrieve additional logs from a running process, or 'ctrl+c' to interrupt.",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "(optional) Session ID for maintaining state across multiple commands. If not provided, a new session will be created.",
			},
		},
		"required": []string{"command"},
	}
}

func (b *Bash) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return &ToolResult{Error: "command parameter is required"}, nil
	}

	sessionID := "default"
	if sid, ok := args["session_id"].(string); ok && sid != "" {
		sessionID = sid
	}

	// Handle special commands
	if command == "ctrl+c" {
		return b.interruptSession(ctx, sessionID)
	}

	// Get or create session
	session := b.getOrCreateSession(sessionID)
	if session == nil {
		return &ToolResult{Error: "Failed to create bash session"}, nil
	}

	// If command is empty, retrieve additional output
	if command == "" {
		return b.retrieveOutput(ctx, session)
	}

	// Execute command
	return b.runCommand(ctx, session, command)
}

func (b *Bash) getOrCreateSession(sessionID string) *BashSession {
	b.mu.Lock()
	defer b.mu.Unlock()

	session, exists := b.sessions[sessionID]
	if exists && session.started {
		return session
	}

	// Create new session
	cmd := exec.Command("/bin/bash")
	cmd.Env = os.Environ()

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil
	}

	if err := cmd.Start(); err != nil {
		return nil
	}

	session = &BashSession{
		process:     cmd,
		started:     true,
		command:     "/bin/bash",
		outputDelay: 200 * time.Millisecond,
		timeout:     120 * time.Second,
		sentinel:    "<<exit>>",
		stdin:       bufio.NewWriter(stdinPipe),
		stdout:      bufio.NewReader(stdoutPipe),
		stderr:      bufio.NewReader(stderrPipe),
	}

	b.sessions[sessionID] = session
	return session
}

func (b *Bash) runCommand(ctx context.Context, session *BashSession, command string) (*ToolResult, error) {
	if !session.started {
		return &ToolResult{Error: "Session has not started"}, nil
	}

	// Check if process is still running
	if session.process.ProcessState != nil && session.process.ProcessState.Exited() {
		return &ToolResult{
			Error: fmt.Sprintf("bash has exited with returncode %d", session.process.ProcessState.ExitCode()),
		}, nil
	}

	if session.timedOut {
		return &ToolResult{
			Error: fmt.Sprintf("timed out: bash has not returned in %v and must be restarted", session.timeout),
		}, nil
	}

	// Send command with sentinel
	fullCommand := command + "; echo '" + session.sentinel + "'\n"
	if _, err := session.stdin.WriteString(fullCommand); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write command: %v", err)}, nil
	}
	session.stdin.Flush()

	// Read output with timeout
	outputCtx, cancel := context.WithTimeout(ctx, session.timeout)
	defer cancel()

	var output strings.Builder
	done := make(chan bool, 1)
	errChan := make(chan error, 1)

	go func() {
		for {
			select {
			case <-outputCtx.Done():
				return
			default:
				time.Sleep(session.outputDelay)
				// Read available data
				buf := make([]byte, 4096)
				n, err := session.stdout.Read(buf)
				if n > 0 {
					output.Write(buf[:n])
					// Check for sentinel
					outputStr := output.String()
					if strings.Contains(outputStr, session.sentinel) {
						// Remove sentinel and everything after it
						idx := strings.Index(outputStr, session.sentinel)
						output.Reset()
						output.WriteString(outputStr[:idx])
						done <- true
						return
					}
				}
				if err != nil {
					if err.Error() != "EOF" {
						errChan <- err
					}
					return
				}
			}
		}
	}()

	select {
	case <-done:
		// Command completed
		outputStr := strings.TrimSuffix(strings.TrimSpace(output.String()), "\n")
		return &ToolResult{Output: outputStr}, nil
	case err := <-errChan:
		return &ToolResult{Error: fmt.Sprintf("Read error: %v", err)}, nil
	case <-outputCtx.Done():
		session.timedOut = true
		return &ToolResult{
			Error: fmt.Sprintf("Command timed out. Sending SIGINT to the process"),
		}, nil
	}
}

func (b *Bash) retrieveOutput(ctx context.Context, session *BashSession) (*ToolResult, error) {
	if !session.started {
		return &ToolResult{Error: "Session has not started"}, nil
	}

	var output strings.Builder
	for session.stdout.Buffered() > 0 {
		line, err := session.stdout.ReadString('\n')
		if err != nil {
			break
		}
		output.WriteString(line)
	}

	outputStr := output.String()
	if outputStr == "" {
		return &ToolResult{Output: "No additional output available"}, nil
	}

	return &ToolResult{Output: strings.TrimSuffix(outputStr, "\n")}, nil
}

func (b *Bash) interruptSession(ctx context.Context, sessionID string) (*ToolResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	session, exists := b.sessions[sessionID]
	if !exists || !session.started {
		return &ToolResult{Error: "Session not found or not started"}, nil
	}

	if session.process.Process != nil {
		session.process.Process.Signal(os.Interrupt)
	}

	return &ToolResult{Output: "Sent interrupt signal to the process"}, nil
}

func (b *Bash) stopSession(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	session, exists := b.sessions[sessionID]
	if !exists {
		return
	}

	if session.process != nil && session.process.Process != nil {
		session.process.Process.Kill()
		session.process.Wait()
	}

	delete(b.sessions, sessionID)
}
