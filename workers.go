package workers

type Brigade[T any, R any] struct {
	Tasks   TaskChan
	Results ResultChan
	Done    chan bool
	count   int
	do      func(int, T) R
}

type TaskChan chan any
type ResultChan chan any

func NewBrigade[T any, R any](n int, do func(int, T) R) *Brigade[T, R] {
	return &Brigade[T, R]{
		Tasks:   make(TaskChan),
		Results: make(ResultChan),
		Done:    make(chan bool),
		count:   n,
		do:      do,
	}
}

func (w *Brigade[T, R]) Close() {
	for i := 0; i < w.count; i++ {
		w.Done <- true
	}
	close(w.Tasks)
	close(w.Results)
	close(w.Done)
}

func (w *Brigade[T, R]) Start() {
	for i := 0; i < w.count; i++ {
		go w.worker(i)
	}
}

func (w *Brigade[T, R]) worker(id int) {
	for {
		select {
		case task := <-w.Tasks:
			w.Results <- w.do(id, task.(T))
		case <-w.Done:
			return
		}
	}
}
