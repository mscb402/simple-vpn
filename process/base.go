package process

type Process interface {
	Run()
	Shutdown()
}
