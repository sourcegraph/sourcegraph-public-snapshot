package docker

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mounttypes "github.com/docker/docker/api/types/mount"
	"github.com/docker/go-units"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Type constants.
const (
	// TypeBind is the type for mounting host dir.
	MountTypeBind mounttypes.Type = "bind"
	// TypeVolume is the type for remote storage volumes.
	MountTypeVolume mounttypes.Type = "volume"
	// TypeTmpfs is the type for mounting tmpfs.
	MountTypeTmpfs mounttypes.Type = "tmpfs"
)

type MountOptions mounttypes.Mount

func (m MountOptions) String() string {
	var sb strings.Builder

	sb.WriteString("type=")
	sb.WriteString(string(m.Type))

	if m.Source != "" {
		sb.WriteString(",source=")
		sb.WriteString(m.Source)
	}

	sb.WriteString(",target=")
	sb.WriteString(m.Target)

	if m.ReadOnly {
		sb.WriteString(",readonly")
	}

	if m.BindOptions != nil {
		switch {
		case m.Consistency != "":
			sb.WriteString(",consistency=")
			sb.WriteString(string(m.Consistency))
		case m.BindOptions.Propagation != "":
			sb.WriteString(",bind-propagation=")
			sb.WriteString(string(m.BindOptions.Propagation))
		case m.BindOptions.NonRecursive:
			sb.WriteString(",bind-nonrecursive")
		}
	}

	if m.VolumeOptions != nil {
		switch {
		case m.VolumeOptions.NoCopy:
			sb.WriteString(",volume-nocopy")
		case m.VolumeOptions.Labels != nil:
			sb.WriteString(",volume-label=")
			for k, v := range m.VolumeOptions.Labels {
				sb.WriteString(k)
				sb.WriteString("=")
				sb.WriteString(v)
			}
		case m.VolumeOptions.DriverConfig != nil:
			sb.WriteString(",volume-driver=")
			sb.WriteString(m.VolumeOptions.DriverConfig.Name)
			if len(m.VolumeOptions.DriverConfig.Options) > 0 {
				sb.WriteString(",volume-opt=")
				for k, v := range m.VolumeOptions.DriverConfig.Options {
					sb.WriteString(k)
					sb.WriteString("=")
					sb.WriteString(v)
				}
			}
		}
	}

	if m.TmpfsOptions != nil {
		switch {
		case m.TmpfsOptions.SizeBytes > 0:
			sb.WriteString(",tmpfs-size=")
			sb.WriteString(strconv.Itoa(int(m.TmpfsOptions.SizeBytes)))
		case m.TmpfsOptions.Mode.String() != "":
			sb.WriteString(",tmpfs-mode=")
			sb.WriteString(fmt.Sprintf("%04o\n", m.TmpfsOptions.Mode.Perm()))
		}
	}

	return sb.String()
}

// ParseMount parses a mount spec like you'd see using the Docker command-line.
// e.g. docker run --rm --mount type=bind,source=$(pwd),destination=/workspace go:latest
// Copied from https://github.com/docker/cli/blob/v24.0.6/opts/mount.go#L23
func ParseMount(spec string) (*MountOptions, error) {
	mount := &MountOptions{}

	csvReader := csv.NewReader(strings.NewReader(spec))
	fields, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	volumeOptions := func() *mounttypes.VolumeOptions {
		if mount.VolumeOptions == nil {
			mount.VolumeOptions = &mounttypes.VolumeOptions{
				Labels: make(map[string]string),
			}
		}
		if mount.VolumeOptions.DriverConfig == nil {
			mount.VolumeOptions.DriverConfig = &mounttypes.Driver{}
		}
		return mount.VolumeOptions
	}

	bindOptions := func() *mounttypes.BindOptions {
		if mount.BindOptions == nil {
			mount.BindOptions = new(mounttypes.BindOptions)
		}
		return mount.BindOptions
	}

	tmpfsOptions := func() *mounttypes.TmpfsOptions {
		if mount.TmpfsOptions == nil {
			mount.TmpfsOptions = new(mounttypes.TmpfsOptions)
		}
		return mount.TmpfsOptions
	}

	setValueOnMap := func(target map[string]string, value string) {
		k, v, _ := strings.Cut(value, "=")
		if k != "" {
			target[k] = v
		}
	}

	mount.Type = mounttypes.TypeVolume // default to volume mounts
	// Set writable as the default
	for _, field := range fields {
		key, val, ok := strings.Cut(field, "=")

		// TODO(thaJeztah): these options should not be case-insensitive.
		key = strings.ToLower(key)

		if !ok {
			switch key {
			case "readonly", "ro":
				mount.ReadOnly = true
				continue
			case "volume-nocopy":
				volumeOptions().NoCopy = true
				continue
			case "bind-nonrecursive":
				bindOptions().NonRecursive = true
				continue
			default:
				return nil, errors.Newf("invalid field '%s' must be a key=value pair", field)
			}
		}

		switch key {
		case "type":
			mount.Type = mounttypes.Type(strings.ToLower(val))
		case "source", "src":
			mount.Source = val
			if strings.HasPrefix(val, "."+string(filepath.Separator)) || val == "." {
				if abs, err := filepath.Abs(val); err == nil {
					mount.Source = abs
				}
			}
		case "target", "dst", "destination":
			mount.Target = val
		case "readonly", "ro":
			mount.ReadOnly, err = strconv.ParseBool(val)
			if err != nil {
				return nil, errors.Newf("invalid value for %s: %s", key, val)
			}
		case "consistency":
			mount.Consistency = mounttypes.Consistency(strings.ToLower(val))
		case "bind-propagation":
			bindOptions().Propagation = mounttypes.Propagation(strings.ToLower(val))
		case "bind-nonrecursive":
			bindOptions().NonRecursive, err = strconv.ParseBool(val)
			if err != nil {
				return nil, errors.Newf("invalid value for %s: %s", key, val)
			}
		case "volume-nocopy":
			volumeOptions().NoCopy, err = strconv.ParseBool(val)
			if err != nil {
				return nil, errors.Newf("invalid value for volume-nocopy: %s", val)
			}
		case "volume-label":
			setValueOnMap(volumeOptions().Labels, val)
		case "volume-driver":
			volumeOptions().DriverConfig.Name = val
		case "volume-opt":
			if volumeOptions().DriverConfig.Options == nil {
				volumeOptions().DriverConfig.Options = make(map[string]string)
			}
			setValueOnMap(volumeOptions().DriverConfig.Options, val)
		case "tmpfs-size":
			sizeBytes, err := units.RAMInBytes(val)
			if err != nil {
				return nil, errors.Newf("invalid value for %s: %s", key, val)
			}
			tmpfsOptions().SizeBytes = sizeBytes
		case "tmpfs-mode":
			ui64, err := strconv.ParseUint(val, 8, 32)
			if err != nil {
				return nil, errors.Newf("invalid value for %s: %s", key, val)
			}
			tmpfsOptions().Mode = os.FileMode(ui64)
		default:
			return nil, errors.Newf("unexpected key '%s' in '%s'", key, field)
		}
	}

	if mount.Type == "" {
		return nil, errors.New("type is required")
	}

	if mount.Target == "" {
		return nil, errors.New("target is required")
	}

	if mount.VolumeOptions != nil && mount.Type != mounttypes.TypeVolume {
		return nil, errors.Newf("cannot mix 'volume-*' options with mount type '%s'", mount.Type)
	}

	if mount.BindOptions != nil && mount.Type != mounttypes.TypeBind {
		return nil, errors.Newf("cannot mix 'bind-*' options with mount type '%s'", mount.Type)
	}

	if mount.TmpfsOptions != nil && mount.Type != mounttypes.TypeTmpfs {
		return nil, errors.Newf("cannot mix 'tmpfs-*' options with mount type '%s'", mount.Type)
	}

	return mount, nil
}
