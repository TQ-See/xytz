package utils

import (
	"log"
	"os/exec"
	"sync"
)

type SearchManager struct {
	cmd      *exec.Cmd
	mutex    sync.Mutex
	canceled bool
}

func NewSearchManager() *SearchManager {
	return &SearchManager{}
}

func (sm *SearchManager) SetCmd(cmd *exec.Cmd) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.cmd = cmd
}

func (sm *SearchManager) GetCmd() *exec.Cmd {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	return sm.cmd
}

func (sm *SearchManager) SetCanceled(canceled bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.canceled = canceled
}

func (sm *SearchManager) WasCanceled() bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	return sm.canceled
}

func (sm *SearchManager) Clear() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.cmd = nil
	sm.canceled = false
}

func (sm *SearchManager) Cancel() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.cmd != nil && sm.cmd.Process != nil {
		sm.canceled = true
		if err := sm.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill search process: %v", err)
			return err
		}
	}

	return nil
}
