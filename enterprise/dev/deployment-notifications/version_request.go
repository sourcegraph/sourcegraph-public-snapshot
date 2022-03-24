package main

type ApplicationVersionDiff struct {
	Old string
	New string
}

type DeploymentDiffer interface {
	Applications() (map[string]*ApplicationVersionDiff, error)
}
