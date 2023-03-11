package lsifstore

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetPackageInformation returns package information data by identifier.
func (s *store) GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	_, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	if strings.HasPrefix(packageInformationID, "scip:") {
		packageInfo := strings.Split(packageInformationID, ":")
		if len(packageInfo) != 4 {
			return precise.PackageInformationData{}, false, errors.Newf("invalid package information ID %q", packageInformationID)
		}

		manager, err := base64.RawStdEncoding.DecodeString(packageInfo[1])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		name, err := base64.RawStdEncoding.DecodeString(packageInfo[2])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		version, err := base64.RawStdEncoding.DecodeString(packageInfo[3])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}

		return precise.PackageInformationData{
			Manager: string(manager),
			Name:    string(name),
			Version: string(version),
		}, true, nil
	}

	return precise.PackageInformationData{}, false, nil
}
