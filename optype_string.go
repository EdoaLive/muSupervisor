// Code generated by "stringer -type opType"; DO NOT EDIT.

package muSupervisor

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LOCK-0]
	_ = x[UNLOCK-1]
	_ = x[RLOCK-2]
	_ = x[RUNLOCK-3]
}

const _opType_name = "LOCKUNLOCKRLOCKRUNLOCK"

var _opType_index = [...]uint8{0, 4, 10, 15, 22}

func (i opType) String() string {
	if i < 0 || i >= opType(len(_opType_index)-1) {
		return "opType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _opType_name[_opType_index[i]:_opType_index[i+1]]
}
