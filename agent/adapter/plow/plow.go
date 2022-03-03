package plow

import (
	"bytes"
	"context"
	"log"
	"net"
	"os/exec"
	"strconv"
)

type PlowJob struct {
	TargetURL   string
	Connections int

	cancel   func()
	finished chan struct{}
}

func NewPlowJob(targetURL string, connections int) *PlowJob {
	return &PlowJob{
		TargetURL:   targetURL,
		Connections: connections,
		finished:    make(chan struct{}, 1),
	}
}

// Process - test process function
func (t *PlowJob) Process() {
	defer func() { t.finished <- struct{}{} }()

	execCtx, execCancel := context.WithCancel(context.Background())
	t.cancel = execCancel

	port, err := getFreePort()
	if err != nil {
		log.Printf("[ERROR] failed to start the Plow [target=%s], "+
			"unable to get a free WebUI port: %s\n", t.TargetURL, err)
		return
	}

	log.Printf("starting the job for the target '%s' [track progress at: http://127.0.0.1:%d ]!\n",
		t.TargetURL, port)

	plowCmd := exec.CommandContext(execCtx,
		"plow",
		t.TargetURL,
		"-c", strconv.Itoa(t.Connections),
		"--listen", ":"+strconv.Itoa(port))

	var errorBuffer = bytes.Buffer{}
	plowCmd.Stderr = &errorBuffer

	err = plowCmd.Run()
	if err != nil && !(execCtx.Err() != nil) {
		log.Printf("[ERROR] failed to start the Plow [target=%s]: %s, [stderr%s]",
			t.TargetURL, err, errorBuffer.String())
		return
	}

	log.Printf("job for is finished'%s'!\n", t.TargetURL)
}

func (t *PlowJob) Cancel() {
	log.Printf("cancelling the job for the Target '%s'\n", t.TargetURL)
	if t.cancel != nil {
		t.cancel()
	}
	log.Printf("job canceled for the Target '%s'\n", t.TargetURL)
}

func (t *PlowJob) FinishedSignal() chan struct{} {
	return t.finished
}

func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
