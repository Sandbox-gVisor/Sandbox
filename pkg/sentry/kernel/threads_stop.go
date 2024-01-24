package kernel

func (t *Task) stopOtherThreadsInTg() {
	if t == nil {
		return
	}
	selfTID := t.ThreadID()

	t.tg.pidns.owner.mu.RLock()
	defer t.tg.pidns.owner.mu.RUnlock()
	t.tg.signalHandlers.mu.Lock()
	defer t.tg.signalHandlers.mu.Unlock()

	for thread := t.tg.tasks.Front(); thread != nil; thread = thread.Next() {
		if selfTID != t.tg.pidns.tids[thread] {
			thread.beginStopLocked()
			thread.interrupt()
		}
	}
}

func (t *Task) resumeOtherThreadsInTg() {
	if t == nil {
		return
	}
	selfTID := t.ThreadID()

	t.tg.pidns.owner.mu.RLock()
	defer t.tg.pidns.owner.mu.RUnlock()
	t.tg.signalHandlers.mu.Lock()
	defer t.tg.signalHandlers.mu.Unlock()

	for thread := t.tg.tasks.Front(); thread != nil; thread = thread.Next() {
		if selfTID != t.tg.pidns.tids[thread] && thread.stopCount.Load() != 0 {
			thread.endStopLocked()
		}
	}
}
