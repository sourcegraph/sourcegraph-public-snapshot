package workerutil

type TestRecord struct {
	ID    int
	State string
}

func (v TestRecord) RecordID() int {
	return v.ID
}
