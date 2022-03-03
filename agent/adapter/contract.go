package adapter

// BenchmarkJob - interface for job processing
type BenchmarkJob interface {
	Process()
	Cancel()
	FinishedSignal() chan struct{}
}
