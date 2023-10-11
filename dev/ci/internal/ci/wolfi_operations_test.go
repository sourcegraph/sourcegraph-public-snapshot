package ci

import (
	"reflect"
	"testing"
)

func Test_sanitizeStepKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			"Test 1",
			"foo!@£_bar$%^baz;'-bam",
			"foo_barbaz-bam",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeStepKey(tt.key); got != tt.want {
				t.Errorf("sanitizeStepKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllImageDependencies(t *testing.T) {
	type args struct {
		wolfiImageDir string
	}
	tests := []struct {
		name                string
		wolfiImageDir       string
		wantPackagesByImage map[string][]string
		wantErr             bool
	}{
		{
			"Test 1",
			"test/wolfi-images",
			map[string][]string{
				"wolfi-test-image-1": {
					"tini",
					"mailcap",
					"git",
					"wolfi-test-package@sourcegraph",
					"wolfi-test-package-subpackage@sourcegraph",
					"foobar-package",
				},
				"wolfi-test-image-2": {
					"tini",
					"mailcap",
					"git",
					"foobar-package",
					"wolfi-test-package-subpackage@sourcegraph",
					"wolfi-test-package-2@sourcegraph",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wolfiImageDirPath := tt.wolfiImageDir
			gotPackagesByImage, err := GetAllImageDependencies(wolfiImageDirPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllImageDependencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPackagesByImage, tt.wantPackagesByImage) {
				t.Errorf("GetAllImageDependencies() = %v, want %v", gotPackagesByImage, tt.wantPackagesByImage)
			}
		})
	}
}

func TestGetDependenciesOfPackage(t *testing.T) {
	type args struct {
		packageName string
		repo        string
	}
	tests := []struct {
		name       string
		args       args
		wantImages []string
	}{
		{
			"Test wolfi-test-package and subpackage",
			args{packageName: "wolfi-test-package", repo: "sourcegraph"},
			[]string{"wolfi-test-image-1", "wolfi-test-image-2"},
		},
		{
			"Test wolfi-test-package-2",
			args{packageName: "wolfi-test-package-2", repo: "sourcegraph"},
			[]string{"wolfi-test-image-2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wolfiImageDirPath := "test/wolfi-images"
			gotPackagesByImage, err := GetAllImageDependencies(wolfiImageDirPath)
			if err != nil {
				t.Errorf("Error running GetAllImageDependencies() error = %v", err)
				return
			}

			if gotImages := GetDependenciesOfPackage(gotPackagesByImage, tt.args.packageName, tt.args.repo); !reflect.DeepEqual(gotImages, tt.wantImages) {
				t.Errorf("GetDependenciesOfPackage() = %v, want %v", gotImages, tt.wantImages)
			}
		})
	}
}
