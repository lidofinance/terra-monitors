package config

import "time"

func FlowManager(interval uint) chan struct{} {
	ch := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Duration(interval) * time.Second):
				ch <- struct{}{}
			}
		}
	}()
	return ch
}
