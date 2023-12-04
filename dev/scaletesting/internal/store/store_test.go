package store

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNew(t *testing.T) {
	c := qt.New(t)
	c.Run("OK", func(c *qt.C) {
		s, cleanup, err := newTestStore(c)
		defer cleanup()
		c.Assert(err, qt.IsNil)
		c.Assert(s, qt.IsNotNil)
	})
}

func TestInsert(t *testing.T) {
	c := qt.New(t)
	s, cleanup, err := newTestStore(c)
	c.Assert(err, qt.IsNil)
	defer cleanup()

	repo1 := Repo{
		Name:     "repo1",
		GitURL:   "https://fromgiturl2",
		ToGitURL: "https://togiturl1",
		Created:  true,
		Pushed:   true,
		Failed:   "err2",
	}
	repo2 := Repo{
		Name:     "repo2",
		GitURL:   "https://fromgiturl2",
		ToGitURL: "https://togiturl2",
		Created:  true,
		Pushed:   true,
		Failed:   "err2",
	}

	repos := []*Repo{&repo1, &repo2}
	err = s.Insert(repos)
	c.Assert(err, qt.IsNil)

	res, err := s.Load()
	c.Assert(err, qt.IsNil)
	c.Assert(res, qt.HasLen, 2)

	for i, got := range res {
		c.Assert(got.Name, qt.Equals, repos[i].Name)
		c.Assert(got.GitURL, qt.Equals, repos[i].GitURL)
		c.Assert(got.ToGitURL, qt.Equals, repos[i].ToGitURL)
		c.Assert(got.Created, qt.Equals, repos[i].Created)
		c.Assert(got.Pushed, qt.Equals, repos[i].Pushed)
		c.Assert(got.Failed, qt.Equals, repos[i].Failed)
	}
}

func TestLoadAndSave(t *testing.T) {
	c := qt.New(t)
	s, cleanup, err := newTestStore(c)
	c.Assert(err, qt.IsNil)
	defer cleanup()

	repo1 := Repo{
		Name:   "repo1",
		GitURL: "https://fromgiturl2",
	}
	repo2 := Repo{
		Name:   "repo2",
		GitURL: "https://fromgiturl1",
	}

	repos := []*Repo{&repo1, &repo2}
	err = s.Insert(repos)
	c.Assert(err, qt.IsNil)

	repo1.Failed = "err1"
	repo1.Created = true
	repo1.Pushed = true
	repo1.ToGitURL = "https://togiturl1"
	err = s.SaveRepo(&repo1)
	c.Assert(err, qt.IsNil)

	repo2.Failed = "err2"
	repo2.Created = true
	repo2.Pushed = true
	repo2.ToGitURL = "https://togiturl2"
	err = s.SaveRepo(&repo2)
	c.Assert(err, qt.IsNil)

	res, err := s.Load()
	c.Assert(err, qt.IsNil)

	for i, got := range res {
		c.Assert(got.Name, qt.Equals, repos[i].Name)
		c.Assert(got.GitURL, qt.Equals, repos[i].GitURL)
		c.Assert(got.ToGitURL, qt.Equals, repos[i].ToGitURL)
		c.Assert(got.Created, qt.Equals, repos[i].Created)
		c.Assert(got.Pushed, qt.Equals, repos[i].Pushed)
	}
}

func newTestStore(c *qt.C) (*Store, func(), error) {
	dir, err := os.MkdirTemp(os.TempDir(), "testdb")
	c.Assert(err, qt.IsNil)
	path := filepath.Join(dir, "test.db")
	teardown := func() {
		_ = os.RemoveAll(path)
	}
	s, err := New(path)
	return s, teardown, err
}
