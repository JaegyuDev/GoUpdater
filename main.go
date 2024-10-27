package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

type Metadata struct {
	FormatVersion int `json:"format_version"`
	Repository    struct {
		Tag      string `json:"tag"`
		Location string `json:"location"`
	} `json:"repository"`
}

type Server struct {
	JavaPath string   `json:"java_path"`
	Jar      string   `json:"jar"`
	Args     []string `json:"args"`
}

type Config struct {
	Metadata Metadata `json:"metadata"`
	Server   Server   `json:"server"`
}

// Function to parse JSON file into Config struct
func loadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}

func checkForUpdates(cfg *Config) {
	os.Getenv()
}

// injectCmd allows finer control over the order of arguments passed to the cmd struct. This is mainly used so that
// you can define things such as `-Xmx` before `-jar`.
func injectCmd(cmd string, jvmArgs []string, args ...string) *exec.Cmd {
	jvmArgs = append(jvmArgs, args...)
	return exec.Command(cmd, jvmArgs...)
}

// TODO: clean up this function
func startServer(cfg *Config) error {
	// Prepare command
	javaPath := cfg.Server.JavaPath
	jvmArgs := cfg.Server.Args
	jarArg := cfg.Server.Jar
	cmd := injectCmd(javaPath, jvmArgs, "-jar", jarArg)

	// Set up pipes for stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating stdin pipe:", err)
		os.Exit(1)
	}

	// Set up a log file for stderr-only logging
	logFile, err := os.OpenFile("session_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error creating log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Start server and tee stderr to both socket and log file
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating stderr pipe:", err)
		os.Exit(1)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating stdout pipe:", err)
		os.Exit(1)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}

	// Set up a Unix socket for admin commands
	socketPath := "/tmp/minecraft_server.sock"
	err = os.Remove(socketPath)
	if err != nil {
		fmt.Println("Error closing socket:", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error creating Unix socket:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on Unix socket:", socketPath)

	// Goroutine to copy stdout to socket
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			// Pipe both stdout and stderr to the socket connection
			go func(c net.Conn) {
				defer c.Close()

				// Forward stdout to connection
				_, err := io.Copy(c, stdout)
				if err != nil {
					fmt.Println("Error writing stdout to socket:", err)
				}
			}(conn)

			// Tee stderr to both log file and socket connection
			go func(c net.Conn) {
				defer c.Close()

				tee := io.TeeReader(stderr, logFile)
				_, err := io.Copy(c, tee)
				if err != nil {
					fmt.Println("Error writing stderr to socket and log:", err)
				}
			}(conn)

			// Forward commands from socket to server stdin
			go func(c net.Conn) {
				defer c.Close()
				_, err := io.Copy(stdin, c)
				if err != nil {
					fmt.Println("Error forwarding commands to server stdin:", err)
				}
			}(conn)
		}
	}()

	// Wait for the command to complete
	return cmd.Wait()
}

func main() {
	// Load configuration
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// this is blocking on the main thread
	if err := startServer(config); err != nil {
		fmt.Println("Server process exited with error:", err)
	}
}
