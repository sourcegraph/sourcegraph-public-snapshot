package client

import (
	clienttypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client_types_types"
)

//
// TODO - get rid of all of this
//

var ErrNotFound = clienttypes.ErrNotFound

type Location = clienttypes.Location
type Range = clienttypes.Range
type Position = clienttypes.Position
type MonikerData = clienttypes.MonikerData
type PackageInformationData = clienttypes.PackageInformationData
type Diagnostic = clienttypes.Diagnostic
type CodeIntelligenceRange = clienttypes.CodeIntelligenceRange
