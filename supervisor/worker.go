package supervisor

import (
	"sync"
	"time"

	"github.com/docker/containerd/runtime"
)

type Worker interface {
	Start()
}

type StartTask struct {
	Container     runtime.Container
	Checkpoint    string
	IO            *runtime.IO
	Stdin         string
	Stdout        string
	Stderr        string
	Err           chan error
	StartResponse chan StartResponse
}

func NewWorker(s *Supervisor, wg *sync.WaitGroup) Worker {
	return &worker{
		s:  s,
		wg: wg,
	}
}

type worker struct {
	wg *sync.WaitGroup
	s  *Supervisor
}

func (w *worker) Start() {
	defer w.wg.Done()
	for t := range w.s.tasks {
		var (
			err     error
			process runtime.Process
			started = time.Now()
		)
		if t.Checkpoint != "" {
			/*
				if err := t.Container.Restore(t.Checkpoint); err != nil {
					evt := NewEvent(DeleteEventType)
					evt.ID = t.Container.ID()
					w.s.SendEvent(evt)
					t.Err <- err
					continue
				}
			*/
		} else {
			process, err = t.Container.Start()
			if err != nil {
				evt := NewEvent(DeleteEventType)
				evt.ID = t.Container.ID()
				w.s.SendEvent(evt)
				t.Err <- err
				continue
			}
		}
		/*
		   if w.s.notifier != nil {
		       n, err := t.Container.OOM()
		       if err != nil {
		           logrus.WithField("error", err).Error("containerd: notify OOM events")
		       } else {
		           w.s.notifier.Add(n, t.Container.ID())
		       }
		   }
		*/
		ContainerStartTimer.UpdateSince(started)
		t.Err <- nil
		t.StartResponse <- StartResponse{
			Stdin:  process.Stdin().Name(),
			Stdout: process.Stdout().Name(),
			Stderr: process.Stderr().Name(),
		}
	}
}
