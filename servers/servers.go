package servers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Amqp-prtcl/snowflakes"
)

type ServerState string

const (
	Starting ServerState = "STARTING"
	Running  ServerState = "RUNNING"
	Stopping ServerState = "STOPPING"
	Closed   ServerState = "CLOSED"
)

var (
	//[20:41:32] [Server thread/INFO]: Stopping server
	stoppingReg = regexp.MustCompile(`\[.+:.+:.+\] \[Server thread\/INFO\]: Stopping server`)

	//[20:41:05] [Server thread/INFO]: Done (14.132s)! For help, type "help"
	RunningReg = regexp.MustCompile(`\[.+:.+:.+\] \[Server thread\/INFO\]: Done \(.*\)! For help, type "help"`)

	ErrAlreadyStarted = fmt.Errorf("Server already started")
)

type Server struct {
	JarPath string
	ID      snowflakes.ID
	State   ServerState

	OnStateChange func(*Server)
	OnLog         func(*Server, string)

	cmd    *exec.Cmd
	input  io.WriteCloser
	output io.ReadCloser

	res    chan error
	logs   chan string
	inputs chan string
}

func NewServer(jarPath string, id snowflakes.ID) *Server {
	return &Server{
		JarPath: jarPath,
		ID:      id,
		State:   Closed,
		inputs:  make(chan string, 10),
	}
}

func (s *Server) setState(st ServerState) {
	s.State = st
	s.OnStateChange(s)
}

func (s *Server) Start() error {
	if s.State != Closed {
		return ErrAlreadyStarted
	}
	s.res = make(chan error, 1)
	s.logs = make(chan string, 10)
	s.inputs = make(chan string, 10)

	s.cmd = exec.Command("java", "-Xmx4G", "-jar", s.JarPath, "nogui")
	s.cmd.Dir = filepath.Dir(s.JarPath)

	var err error
	s.input, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}
	s.output, err = s.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	s.cmd.Stderr = os.Stderr

	go s.listenServer()
	err = s.cmd.Start()
	if err != nil {
		return err
	}

	s.setState(Starting)

	go s.processHandler()
	go func() {
		s.res <- s.cmd.Wait()
	}()
	return nil
}

func (s *Server) SendCommand(cmd string) {
	if !strings.HasSuffix(cmd, "\n") {
		cmd += "\n"
	}
	s.inputs <- cmd
}

func (s *Server) Stop() {
	s.SendCommand("stop")
}

func (s *Server) processHandler() {
	for {
		select {
		case log := <-s.logs:
			s.OnLog(s, log)
			switch s.State {
			case Starting:
				if RunningReg.MatchString(log) {
					s.setState(Running)
				}
			case Running:
				if stoppingReg.MatchString(log) {
					s.setState(Stopping)
				}
			}

		case err := <-s.res:
			s.setState(Closed)
			if err != nil {
				panic(err)
			}
			s.inputs = nil
			s.logs = nil
			s.res = nil
			return

		case in := <-s.inputs:
			_, err := s.input.Write([]byte(in))
			if err != nil && err != io.EOF {
				panic(err)
			}
		}
	}
}

func (s *Server) listenServer() {
	r := bufio.NewReader(s.output)
	var str string
	var err error
	for {
		str, err = r.ReadString('\n')
		if err != nil { // should be only EOF
			return
		}
		s.logs <- str
	}
}
