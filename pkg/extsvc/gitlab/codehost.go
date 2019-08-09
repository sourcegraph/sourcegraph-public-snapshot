package gitlab

import (
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

// ServiceType is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID
// value is the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
const ServiceType = "gitlab"

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitLab project.
func ExternalRepoSpec(proj *Project, baseURL url.URL) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ID:          strconv.Itoa(proj.ID),
		ServiceType: ServiceType,
		ServiceID:   extsvc.NormalizeBaseURL(&baseURL).String(),
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_802(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
