package graphstoreutil

import "testing"

func TestEvenlyDistributedRepoPaths(t *testing.T) {
	rp := &EvenlyDistributedRepoPaths{}

	tests := []string{
		"example.com/my/repo",
		"foo.com/my/repo",
	}
	for _, testRepo := range tests {
		path := rp.RepoToPath(testRepo)
		repo := rp.PathToRepo(path)

		t.Logf("%s -> %v (len %d)", testRepo, path, len(path[0]))

		if repo != testRepo {
			t.Errorf("got %q, want %q", repo, testRepo)
		}
	}
}

func BenchmarkEvenlyDistributedRepoPaths_RepoToPath(b *testing.B) {
	rp := &EvenlyDistributedRepoPaths{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rp.RepoToPath("example.com/my/repo")
	}
}

func BenchmarkEvenlyDistributedRepoPaths_PathToRepo(b *testing.B) {
	rp := &EvenlyDistributedRepoPaths{}
	path := rp.RepoToPath("example.com/my/repo")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rp.PathToRepo(path)
	}
}
