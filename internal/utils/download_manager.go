package utils

import (
	"context"
	"log"
	"os/exec"
	"sync"
)

type DownloadManager struct {
	cmd      *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
	mutex    sync.Mutex
	isPaused bool
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{}
}

func (dm *DownloadManager) SetCmd(cmd *exec.Cmd) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.cmd = cmd
}

func (dm *DownloadManager) GetCmd() *exec.Cmd {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	return dm.cmd
}

func (dm *DownloadManager) SetContext(ctx context.Context, cancel context.CancelFunc) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.ctx = ctx
	dm.cancel = cancel
}

func (dm *DownloadManager) GetContext() (context.Context, context.CancelFunc) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	return dm.ctx, dm.cancel
}

func (dm *DownloadManager) SetPaused(paused bool) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.isPaused = paused
}

func (dm *DownloadManager) IsPaused() bool {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	return dm.isPaused
}

func (dm *DownloadManager) Clear() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.cmd = nil
	dm.ctx = nil
	dm.cancel = nil
	dm.isPaused = false
}

func (dm *DownloadManager) Cancel() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.cancel != nil {
		dm.cancel()
	}

	if dm.cmd != nil && dm.cmd.Process != nil {
		if err := dm.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill download process: %v", err)
			return err
		}
	}

	return nil
}
