package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const upsertEmptyBatchChangeQuery = `
mutation UpsertEmptyBatchChange(
	$name: String!
	$namespace: ID!
) {
	upsertEmptyBatchChange(
		name: $name,
		namespace: $namespace
	) {
		id
		name
	}
}
`

func (svc *Service) UpsertBatchChange(
	ctx context.Context,
	name string,
	namespaceID string,
) (string, string, error) {
	var resp struct {
		UpsertEmptyBatchChange struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"upsertEmptyBatchChange"`
	}

	if ok, err := svc.client.NewRequest(upsertEmptyBatchChangeQuery, map[string]interface{}{
		"name":      name,
		"namespace": namespaceID,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", "", err
	}

	return resp.UpsertEmptyBatchChange.ID, resp.UpsertEmptyBatchChange.Name, nil
}

const createBatchSpecFromRawQuery = `
mutation CreateBatchSpecFromRaw(
    $batchSpec: String!,
    $namespace: ID!,
    $allowIgnored: Boolean!,
    $allowUnsupported: Boolean!,
    $noCache: Boolean!,
    $batchChange: ID!,
) {
    createBatchSpecFromRaw(
        batchSpec: $batchSpec,
        namespace: $namespace,
        allowIgnored: $allowIgnored,
        allowUnsupported: $allowUnsupported,
        noCache: $noCache,
        batchChange: $batchChange,
    ) {
        id
    }
}
`

func (svc *Service) CreateBatchSpecFromRaw(
	ctx context.Context,
	batchSpec string,
	namespaceID string,
	allowIgnored bool,
	allowUnsupported bool,
	noCache bool,
	batchChange string,
) (string, error) {
	var resp struct {
		CreateBatchSpecFromRaw struct {
			ID string `json:"id"`
		} `json:"createBatchSpecFromRaw"`
	}

	if ok, err := svc.client.NewRequest(createBatchSpecFromRawQuery, map[string]interface{}{
		"batchSpec":        batchSpec,
		"namespace":        namespaceID,
		"allowIgnored":     allowIgnored,
		"allowUnsupported": allowUnsupported,
		"noCache":          noCache,
		"batchChange":      batchChange,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", err
	}

	return resp.CreateBatchSpecFromRaw.ID, nil
}

// UploadBatchSpecWorkspaceFiles uploads workspace files to the server.
func (svc *Service) UploadBatchSpecWorkspaceFiles(ctx context.Context, workingDir string, batchSpecID string, steps []batches.Step) error {
	filePaths := make(map[string]bool)
	for _, step := range steps {
		for _, mount := range step.Mount {
			paths, err := getFilePaths(workingDir, mount.Path)
			if err != nil {
				return err
			}
			// Dedupe any files.
			for _, path := range paths {
				if !filePaths[path] {
					filePaths[path] = true
				}
			}
		}
	}

	for filePath := range filePaths {
		if err := svc.uploadFile(ctx, workingDir, filePath, batchSpecID); err != nil {
			return err
		}
	}
	return nil
}

func getFilePaths(workingDir, filePath string) ([]string, error) {
	var filePaths []string
	actualFilePath := filepath.Join(workingDir, filePath)
	info, err := os.Stat(actualFilePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		dir, err := os.ReadDir(actualFilePath)
		if err != nil {
			return nil, err
		}
		for _, dirEntry := range dir {
			paths, err := getFilePaths(workingDir, filepath.Join(filePath, dirEntry.Name()))
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, paths...)
		}
	} else {
		relPath, err := filepath.Rel(workingDir, actualFilePath)
		if err != nil {
			return nil, err
		}
		filePaths = append(filePaths, relPath)
	}
	return filePaths, nil
}

func (svc *Service) uploadFile(ctx context.Context, workingDir, filePath, batchSpecID string) error {
	// Create a pipe so the requests can be chunked to the server
	pipeReader, pipeWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(pipeWriter)

	// Write in a separate goroutine to properly chunk the file content. Writing to the pipe lets us not have
	// to put the whole file in memory.
	go func() {
		defer pipeWriter.Close()
		defer multipartWriter.Close()

		if err := createFormFile(multipartWriter, workingDir, filePath); err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	request, err := svc.client.NewHTTPRequest(ctx, http.MethodPost, fmt.Sprintf(".api/files/batch-changes/%s", batchSpecID), pipeReader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	resp, err := svc.client.Do(request)
	if err != nil {
		// Errors passed to pipeWriter.CloseWithError come through here.
		return err
	}
	defer resp.Body.Close()

	// 2xx and 3xx are ok
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		p, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(p))
	}
	return nil
}

const maxFileSize = 10 << 20 // 10MB

func createFormFile(w *multipart.Writer, workingDir string, mountPath string) error {
	f, err := os.Open(filepath.Join(workingDir, mountPath))
	if err != nil {
		return err
	}
	defer f.Close()

	// Limit the size of file to 10MB
	fileStat, err := f.Stat()
	if err != nil {
		return err
	}
	if fileStat.Size() > maxFileSize {
		return errors.Newf("file %q exceeds limit of 10MB", mountPath)
	}

	filePath, fileName := filepath.Split(mountPath)
	filePath = strings.Trim(strings.TrimSuffix(filePath, string(filepath.Separator)), ".")
	// Ensure Windows separators are changed to Unix.
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	if err = w.WriteField("filepath", filePath); err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	if err = w.WriteField("filemod", fileInfo.ModTime().UTC().String()); err != nil {
		return err
	}

	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, f); err != nil {
		return err
	}
	return nil
}

const executeBatchSpecQuery = `
mutation ExecuteBatchSpec($batchSpec: ID!, $noCache: Boolean!) {
    executeBatchSpec(batchSpec: $batchSpec, noCache: $noCache) {
        id
    }
}
`

func (svc *Service) ExecuteBatchSpec(
	ctx context.Context,
	batchSpecID string,
	noCache bool,
) (string, error) {
	var resp struct {
		ExecuteBatchSpec struct {
			ID string `json:"id"`
		} `json:"executeBatchSpec"`
	}

	if ok, err := svc.client.NewRequest(executeBatchSpecQuery, map[string]interface{}{
		"batchSpec": batchSpecID,
		"noCache":   noCache,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", err
	}

	return resp.ExecuteBatchSpec.ID, nil
}

const batchSpecWorkspaceResolutionQuery = `
query BatchSpecWorkspaceResolution($batchSpec: ID!) {
    node(id: $batchSpec) {
        ... on BatchSpec {
            workspaceResolution {
                failureMessage
                state
                workspaces {
                    totalCount
                }
            }
        }
    }
}
`

type BatchSpecWorkspaceResolution struct {
	FailureMessage string `json:"failureMessage"`
	State          string `json:"state"`
	Workspaces     struct {
		TotalCount int `json:"totalCount"`
	} `json:"workspaces"`
}

func (svc *Service) GetBatchSpecWorkspaceResolution(ctx context.Context, id string) (*BatchSpecWorkspaceResolution, error) {
	var resp struct {
		Node struct {
			WorkspaceResolution BatchSpecWorkspaceResolution `json:"workspaceResolution"`
		} `json:"node"`
	}

	if ok, err := svc.client.NewRequest(batchSpecWorkspaceResolutionQuery, map[string]interface{}{
		"batchSpec": id,
	}).Do(ctx, &resp); err != nil || !ok {
		return nil, err
	}

	return &resp.Node.WorkspaceResolution, nil
}
