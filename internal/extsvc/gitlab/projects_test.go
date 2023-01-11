package gitlab

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

// TestClient_GetProject tests the behavior of GetProject.
func TestClient_GetProject(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
	"id": 1,
	"path_with_namespace": "n1/n2/r",
	"description": "d",
	"forks_count": 1,
	"star_count": 100,
	"web_url": "https://gitlab.example.com/n1/n2/r",
	"http_url_to_repo": "https://gitlab.example.com/n1/n2/r.git",
	"ssh_url_to_repo": "git@gitlab.example.com:n1/n2/r.git"
}
`,
	}
	c := newTestClient(t)
	c.httpClient = &mock

	want := Project{
		ForksCount: 1,
		StarCount:  100,
		ProjectCommon: ProjectCommon{
			ID:                1,
			PathWithNamespace: "n1/n2/r",
			Description:       "d",
			WebURL:            "https://gitlab.example.com/n1/n2/r",
			HTTPURLToRepo:     "https://gitlab.example.com/n1/n2/r.git",
			SSHURLToRepo:      "git@gitlab.example.com:n1/n2/r.git",
		},
	}

	// Test first fetch (cache empty)
	proj, err := c.GetProject(context.Background(), GetProjectOp{PathWithNamespace: "n1/n2/r"})
	if err != nil {
		t.Fatal(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cache once", mock.count)
	}
	if !reflect.DeepEqual(proj, &want) {
		t.Errorf("got project %+v, want %+v", proj, &want)
	}

	// Test that proj is cached (and therefore NOT fetched) from client on second request.
	proj, err = c.GetProject(context.Background(), GetProjectOp{PathWithNamespace: "n1/n2/r"})
	if err != nil {
		t.Fatal(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to hit cache", mock.count)
	}
	if !reflect.DeepEqual(proj, &want) {
		t.Errorf("got project %+v, want %+v", proj, &want)
	}

	// Test the `NoCache: true` option
	proj, err = c.GetProject(context.Background(), GetProjectOp{PathWithNamespace: "n1/n2/r", CommonOp: CommonOp{NoCache: true}})
	if err != nil {
		t.Fatal(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 2 {
		t.Errorf("mock.count == %d, expected to hit cache", mock.count)
	}
	if !reflect.DeepEqual(proj, &want) {
		t.Errorf("got project %+v, want %+v", proj, &want)
	}
}

// TestClient_GetProject_nonexistent tests the behavior of GetProject when called
// on a project that does not exist.
func TestClient_GetProject_nonexistent(t *testing.T) {
	mock := mockHTTPEmptyResponse{http.StatusNotFound}
	c := newTestClient(t)
	c.httpClient = &mock

	proj, err := c.GetProject(context.Background(), GetProjectOp{PathWithNamespace: "doesnt/exist"})
	if !IsNotFound(err) {
		t.Errorf("got err == %v, want IsNotFound(err) == true", err)
	}
	if !errcode.IsNotFound(err) {
		t.Errorf("expected a not found error")
	}
	if proj != nil {
		t.Error("proj != nil")
	}
}

func TestClient_ForkProject(t *testing.T) {
	ctx := context.Background()

	// We'll grab a project to use in the other tests.
	project, err := createTestClient(t).GetProject(ctx, GetProjectOp{
		PathWithNamespace: "sourcegraph/src-cli",
		CommonOp:          CommonOp{NoCache: true},
	})
	assert.Nil(t, err)

	t.Run("success", func(t *testing.T) {
		// For this test to be updated, src-cli must _not_ have been forked into
		// the user associated with $GITLAB_TOKEN.

		name := "sourcegraph-src-cli"
		fork, err := createTestClient(t).ForkProject(ctx, project, nil, name)
		assert.Nil(t, err)
		assert.NotNil(t, fork)

		assert.Nil(t, err)
		forkName, err := fork.Name()
		assert.Nil(t, err)
		assert.Equal(t, name, forkName)
	})

	t.Run("already forked", func(t *testing.T) {
		// For this test to be updated, src-cli must have been forked into the user
		// associated with $GITLAB_TOKEN.
		name := "sourcegraph-src-cli"
		fork, err := createTestClient(t).ForkProject(ctx, project, nil, name)
		assert.Nil(t, err)
		assert.NotNil(t, fork)

		assert.Nil(t, err)
		forkName, err := fork.Name()
		assert.Nil(t, err)
		assert.Equal(t, name, forkName)
	})

	t.Run("error", func(t *testing.T) {
		name := "sourcegraph-src-cli"
		mock := mockHTTPEmptyResponse{http.StatusNotFound}
		c := newTestClient(t)
		c.httpClient = &mock

		fork, err := c.ForkProject(ctx, project, nil, name)
		assert.Nil(t, fork)
		assert.NotNil(t, err)
	})
}

func TestProjectCommon_Name(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for name, pc := range map[string]ProjectCommon{
			"empty":      {PathWithNamespace: ""},
			"no slashes": {PathWithNamespace: "foo"},
		} {
			t.Run(name, func(t *testing.T) {
				name, err := pc.Name()
				assert.Equal(t, "", name)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			pc   ProjectCommon
			want string
		}{
			"single namespace": {
				pc:   ProjectCommon{PathWithNamespace: "foo/bar"},
				want: "bar",
			},
			"nested namespaces": {
				pc:   ProjectCommon{PathWithNamespace: "foo/bar/quux/baz"},
				want: "baz",
			},
		} {
			t.Run(name, func(t *testing.T) {
				name, err := tc.pc.Name()
				assert.Nil(t, err)
				assert.Equal(t, tc.want, name)
			})
		}
	})
}

func TestProjectCommon_Namespace(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for name, pc := range map[string]ProjectCommon{
			"empty":      {PathWithNamespace: ""},
			"no slashes": {PathWithNamespace: "foo"},
		} {
			t.Run(name, func(t *testing.T) {
				ns, err := pc.Namespace()
				assert.Equal(t, "", ns)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			pc   ProjectCommon
			want string
		}{
			"single namespace": {
				pc:   ProjectCommon{PathWithNamespace: "foo/bar"},
				want: "foo",
			},
			"nested namespaces": {
				pc:   ProjectCommon{PathWithNamespace: "foo/bar/quux/baz"},
				want: "foo/bar/quux",
			},
		} {
			t.Run(name, func(t *testing.T) {
				ns, err := tc.pc.Namespace()
				assert.Nil(t, err)
				assert.Equal(t, tc.want, ns)
			})
		}
	})
}
