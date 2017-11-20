package jsoncommentstrip

// Reader reader which reads JSON and strips comments from it
type Reader interface {
	Read(buff []byte) (count int, err error)
}
