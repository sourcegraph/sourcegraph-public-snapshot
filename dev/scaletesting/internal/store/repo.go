package store

// Repo represents a repository, regardless from which codehost it came from.
type Repo struct {
	GitURL   string
	ToGitURL string
	Name     string
	Failed   string
	Created  bool
	Pushed   bool
}
