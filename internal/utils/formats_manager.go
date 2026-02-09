package utils

import (
	"log"
	"os/exec"
	"sync"
)

type FormatsManager struct {
	cmd      *exec.Cmd
	mutex    sync.Mutex
	canceled bool
}

func NewFormatsManager() *FormatsManager {
	return &FormatsManager{}
}

func (fm *FormatsManager) SetCmd(cmd *exec.Cmd) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.cmd = cmd
}

func (fm *FormatsManager) GetCmd() *exec.Cmd {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	return fm.cmd
}

func (fm *FormatsManager) SetCanceled(canceled bool) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.canceled = canceled
}

func (fm *FormatsManager) WasCanceled() bool {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	return fm.canceled
}

func (fm *FormatsManager) Clear() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.cmd = nil
	fm.canceled = false
}

func (fm *FormatsManager) Cancel() error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if fm.cmd != nil && fm.cmd.Process != nil {
		fm.canceled = true
		if err := fm.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill formats process: %v", err)
			return err
		}
	}

	return nil
}
