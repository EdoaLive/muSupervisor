package muSupervisor

import (
	"fmt"
	"time"
)

var opReq chan *opData
var opObtained chan *opData
var unlockChan chan *opData

func init() {
	opReq = make(chan *opData)
	opObtained = make(chan *opData)
	unlockChan = make(chan *opData)

	go supervisor()
}

func supervisor() {
	if Opts.Disable == true {
		fmt.Println("supervisor disabled.")
		return
	}
	fmt.Println("supervisor enabled.")

	newMutexMap := make(mutexOpMap)

	t := time.NewTicker(Opts.CheckFrequency)

	for {
		select {
		case opD := <-opReq:
			opD.reqTime = time.Now()
			newMutexMap.mapRequest(opD)

		case opD := <-opObtained:
			opD.trackObtained()

		case opD := <-unlockChan:
			newMutexMap.mapUnlock(opD)

		case <-t.C:
			newMutexMap.doCheck()
		}

	}
}

func logOpDetails(op *opData) {
	now := time.Now()
	fmt.Printf("\tGoRoutine: %v, state: %v, type: %v, time: %v, request: %p, mutex: %p\n",
		op.numRoutine,
		op.state,
		op.t,
		now.Sub(op.reqTime),
		op,
		op.mutexPtr,
	)
	if op.stackTrace != nil {
		fmt.Printf("%s\n\n", *op.stackTrace)
	}
}
