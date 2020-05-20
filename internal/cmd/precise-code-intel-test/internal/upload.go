package internal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/efritz/pentimento"
)

// Upload represents an uploaded LSIF Index.
type Upload struct {
	Owner    string
	Name     string
	Rev      string
	UploadID string
}

// UploadParallel uploads the index files specified by each repository and revision to
// the frontend. This returns a list of upload IDs returned by the frontend on success.
func UploadParallel(cacheDir, baseURL string, maxConcurrency int, repos []Repo) ([]Upload, error) {
	total := 0
	for _, repo := range repos {
		total += len(repo.Revs)
	}
	uploadCh := make(chan Upload, total)

	var fns []FnPair
	for _, repo := range repos {
		localRepo := repo

		for _, rev := range repo.Revs {
			localRev := rev

			fns = append(fns, FnPair{
				Fn: func() error {
					id, err := upload(cacheDir, baseURL, localRepo.Owner, localRepo.Name, localRev)
					if err != nil {
						return err
					}

					uploadCh <- Upload{
						Owner:    localRepo.Owner,
						Name:     localRepo.Name,
						Rev:      localRev,
						UploadID: id,
					}
					return nil
				},
				Description: fmt.Sprintf("Uploading %s/%s@%s", localRepo.Owner, localRepo.Name, localRev),
			})
		}
	}

	if err := RunParallel(maxConcurrency, fns); err != nil {
		return nil, err
	}
	close(uploadCh)

	var uploads []Upload
	for upload := range uploadCh {
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

// upload sends the LSIF index specified by the repository and revision to the frontend.
// This returns the upload ID returned by the frontend on success.
func upload(cacheDir, baseURL, owner, name, rev string) (string, error) {
	output, err := runCommandOutput(
		filepath.Join(cacheDir, "indexes"),
		"src",
		fmt.Sprintf("-endpoint=%s", baseURL),
		"lsif",
		"upload",
		"-root=/",
		fmt.Sprintf("-repo=%s", fmt.Sprintf("github.com/%s/%s", owner, name)),
		fmt.Sprintf("-commit=%s", rev),
		fmt.Sprintf("-file=%s", filepath.Join(fmt.Sprintf("%s.%s.dump", name, rev))),
	)
	if err != nil {
		return "", err
	}

	pattern, err := regexp.Compile(`/code-intelligence/lsif-uploads/([0-9A-Za-z=]+)\.`)
	if err != nil {
		return "", err
	}

	return string(pattern.FindSubmatch(output)[1]), nil
}

// WaitForSuccessAll waits until all uploads have finished processing. If any upload
// errors during processing, this function returns a non-nil error.
func WaitForSuccessAll(baseURL, token string, uploads []Upload) error {
	return pentimento.PrintProgress(func(p *pentimento.Printer) error {
		for len(uploads) > 0 {
			var ids []string
			for _, upload := range uploads {
				ids = append(ids, upload.UploadID)
			}

			states, err := uploadStates(baseURL, token, ids)
			if err != nil {
				_ = p.Reset()
				return err
			}

			var temp []Upload
			for _, upload := range uploads {
				switch states[upload.UploadID] {
				case "ERRORED":
					_ = p.Reset()
					return fmt.Errorf("upload %s/%s@%s errored", upload.Owner, upload.Name, upload.Rev)
				case "COMPLETED":
					continue
				}

				temp = append(temp, upload)
			}

			content := pentimento.NewContent()
			for _, upload := range uploads {
				content.AddLine(fmt.Sprintf("%s Processing %s/%s@%s: %s", pentimento.Dots, upload.Owner, upload.Name, upload.Rev, states[upload.UploadID]))
			}
			_ = p.WriteContent(content)

			if len(temp) > 0 {
				time.Sleep(time.Millisecond * 500)
			}
			uploads = temp
		}

		_ = p.Reset()
		return nil
	})
}

// uploadStates gets the states for the given upload identifiers.
func uploadStates(baseURL, token string, ids []string) (map[string]string, error) {
	var fragments []string
	for i, id := range ids {
		fragments = append(fragments, fmt.Sprintf(`
			u%d: node(id: "%s") {
				... on LSIFUpload {
					state
				}
			}
		`, i, id))
	}
	query := fmt.Sprintf("{%s}", strings.Join(fragments, "\n"))

	payload := struct {
		Data map[string]struct {
			State string `json:"state"`
		} `json:"data"`
	}{}
	if err := graphQL(baseURL, token, query, nil, &payload); err != nil {
		return nil, err
	}

	states := map[string]string{}
	for i, id := range ids {
		states[id] = payload.Data[fmt.Sprintf("u%d", i)].State
	}

	return states, nil
}
