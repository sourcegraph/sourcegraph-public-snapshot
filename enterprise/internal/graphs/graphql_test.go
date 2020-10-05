package graphs

type apitestGraph struct {
	ID          string
	Owner       apitestGraphOwner
	Name        string
	Description string
	Spec        string
}

type apitestGraphOwner struct {
	ID         string
	DatabaseID int32
	Name       string
}
