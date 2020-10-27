package command

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatFirecrackerCommandRaw(t *testing.T) {
	actual := formatFirecrackerCommand(
		CommandSpec{
			Commands: []string{"ls", "-a"},
			Dir:      "subdir",
			Env:      []string{"TEST=true"},
		},
		"deadbeef",
		"/proj/src",
		Options{},
	)

	expected := command{
		Commands: []string{
			"ignite", "exec", "deadbeef", "--",
			"cd /work/subdir && TEST=true ls -a",
		},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestFormatFirecrackerCommandDocker(t *testing.T) {
	actual := formatFirecrackerCommand(
		CommandSpec{
			Image:    "alpine:latest",
			Commands: []string{"ls", "-a"},
			Dir:      "subdir",
			Env:      []string{"TEST=true"},
		},
		"deadbeef",
		"/proj/src",
		Options{
			ResourceOptions: ResourceOptions{
				NumCPUs: 4,
				Memory:  "20G",
			},
		},
	)

	expected := command{
		Commands: []string{
			"ignite", "exec", "deadbeef", "--",
			strings.Join([]string{
				"docker", "run", "--rm",
				"--cpus", "4",
				"--memory", "20G",
				"-v", "/work:/data",
				"-w", "/data/subdir",
				"-e", "TEST=true",
				"alpine:latest",
				"ls", "-a",
			}, " "),
		},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestSetupFirecracker(t *testing.T) {
	runner := NewMockCommandRunner()
	if err := setupFirecracker(context.Background(), runner, nil, "deadbeef", "/proj", []string{"img1", "img2", "img3"}, Options{
		FirecrackerOptions: FirecrackerOptions{
			Image:             "ignite-ubuntu",
			ImageArchivesPath: "/archives",
		},
		ResourceOptions: ResourceOptions{
			NumCPUs:   4,
			Memory:    "20G",
			DiskSpace: "1T",
		},
	}); err != nil {
		t.Fatalf("unexpected error tearing down virtual machine: %s", err)
	}

	var actual []string
	for _, call := range runner.RunCommandFunc.History() {
		actual = append(actual, strings.Join(call.Arg2.Commands, " "))
	}

	expected := []string{
		"docker pull img1",
		"docker save -o /archives/image0.tar img1",
		"docker pull img2",
		"docker save -o /archives/image1.tar img2",
		"docker pull img3",
		"docker save -o /archives/image2.tar img3",
		strings.Join([]string{
			"ignite run",
			"--runtime docker --network-plugin docker-bridge",
			"--cpus 4 --memory 20G --size 1T",
			"--copy-files /archives/image0.tar:/image0.tar",
			"--copy-files /archives/image1.tar:/image1.tar",
			"--copy-files /archives/image2.tar:/image2.tar",
			"--copy-files /proj:/work",
			"--ssh --name deadbeef",
			"ignite-ubuntu",
		}, " "),
		"ignite exec deadbeef -- docker load -i /image0.tar",
		"ignite exec deadbeef -- docker load -i /image1.tar",
		"ignite exec deadbeef -- docker load -i /image2.tar",
		"ignite exec deadbeef -- rm /image0.tar",
		"ignite exec deadbeef -- rm /image1.tar",
		"ignite exec deadbeef -- rm /image2.tar",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestTeardownFirecracker(t *testing.T) {
	runner := NewMockCommandRunner()
	if err := teardownFirecracker(context.Background(), runner, nil, "deadbeef"); err != nil {
		t.Fatalf("unexpected error tearing down virtual machine: %s", err)
	}

	var actual []string
	for _, call := range runner.RunCommandFunc.History() {
		actual = append(actual, strings.Join(call.Arg2.Commands, " "))
	}

	expected := []string{
		"ignite stop --runtime docker --network-plugin docker-bridge deadbeef",
		"ignite rm -f --runtime docker --network-plugin docker-bridge deadbeef",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestSanitizeImage(t *testing.T) {
	image := "sourcegraph/ignite-ubuntu"
	tag := ":insiders"
	digest := "@sha256:e54a802a8bec44492deee944acc560e4e0a98f6ffa9a5038f0abac1af677e134"

	testCases := map[string]string{
		"":                   "",          // no regex match (no crash)
		image:                image,       // no tag or hash
		image + digest:       image,       // remove hash without tag
		image + tag:          image + tag, // tag only
		image + tag + digest: image + tag, // tag and hash - keep only tag
	}

	for input, expected := range testCases {
		name := fmt.Sprintf("input=%s", input)

		t.Run(name, func(t *testing.T) {
			if image := sanitizeImage(input); image != expected {
				t.Errorf("unexpected image. want=%q have=%q", expected, image)
			}
		})
	}
}
