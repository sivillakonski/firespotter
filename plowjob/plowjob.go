package plowjob

import (
	"bytes"
	"context"
	"log"
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

	log.Printf("starting the job for the target '%s'!\n", t.TargetURL)

	plowCmd := exec.CommandContext(execCtx,
		"plow",
		t.TargetURL,
		"-c", strconv.Itoa(t.Connections),
		"--listen", "")

	var errorBuffer = bytes.Buffer{}
	plowCmd.Stderr = &errorBuffer

	err := plowCmd.Run()
	if err != nil && !(execCtx.Err() != nil) {
		log.Printf("[ERROR] failed to start the Plow [target=%s]: %s\n [stderr%s]",
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
