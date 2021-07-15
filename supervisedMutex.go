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

	switch t {
	case LOCK, RLOCK:

		st := getStackTrace()
		op.stackTrace = &st

		opReq <- &op
		f()
		opObtained <- &op

	case UNLOCK, RUNLOCK:
		f()
		unlockChan <- &op
	}
}
