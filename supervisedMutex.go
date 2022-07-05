package muSupervisor

type supervisedMutex struct{}

func (m *supervisedMutex) mutexOp(t opType, f func()) {

	if Opts.Disable == true {
		f()
		return
	}

	op := opData{
		t:          t,
		numRoutine: routineNum(getGID()),
		mutexPtr:   m,
	}

	//TODO: While lock requested and obtained are always the same op, unlock may be not.
	// There's no need to create an opData for all 3 channels: cleanup this code.

	switch t {
	case LOCK, RLOCK:

		st := getStackTrace()
		op.stackTrace = &st

		opReq <- &op
		f()
		opObtained <- &op

	case UNLOCK, RUNLOCK:
		unlockChan <- &op
		<-unlockedChan // wait for unlock to be done
		f()
	}
}
