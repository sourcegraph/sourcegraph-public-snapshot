// The SystemInfo domain defines methods and events for querying low-level system information. (experimental)
package systeminfo

import (
	"github.com/neelance/cdp-go/rpc"
)

// The SystemInfo domain defines methods and events for querying low-level system information. (experimental)
type Client struct {
	*rpc.Client
}

// Describes a single graphics processor (GPU).

type GPUDevice struct {
	// PCI ID of the GPU vendor, if available; 0 otherwise.
	VendorId float64 `json:"vendorId"`

	// PCI ID of the GPU device, if available; 0 otherwise.
	DeviceId float64 `json:"deviceId"`

	// String description of the GPU vendor, if the PCI ID is not available.
	VendorString string `json:"vendorString"`

	// String description of the GPU device, if the PCI ID is not available.
	DeviceString string `json:"deviceString"`
}

// Provides information about the GPU(s) on the system.

type GPUInfo struct {
	// The graphics devices on the system. Element 0 is the primary GPU.
	Devices []*GPUDevice `json:"devices"`

	// An optional dictionary of additional GPU related attributes. (optional)
	AuxAttributes interface{} `json:"auxAttributes,omitempty"`

	// An optional dictionary of graphics features and their status. (optional)
	FeatureStatus interface{} `json:"featureStatus,omitempty"`

	// An optional array of GPU driver bug workarounds.
	DriverBugWorkarounds []string `json:"driverBugWorkarounds"`
}

type GetInfoRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns information about the system.
func (d *Client) GetInfo() *GetInfoRequest {
	return &GetInfoRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetInfoResult struct {
	// Information about the GPUs on the system.
	Gpu *GPUInfo `json:"gpu"`

	// A platform-dependent description of the model of the machine. On Mac OS, this is, for example, 'MacBookPro'. Will be the empty string if not supported.
	ModelName string `json:"modelName"`

	// A platform-dependent description of the version of the machine. On Mac OS, this is, for example, '10.1'. Will be the empty string if not supported.
	ModelVersion string `json:"modelVersion"`

	// The command line string used to launch the browser. Will be the empty string if not supported.
	CommandLine string `json:"commandLine"`
}

func (r *GetInfoRequest) Do() (*GetInfoResult, error) {
	var result GetInfoResult
	err := r.client.Call("SystemInfo.getInfo", r.opts, &result)
	return &result, err
}

func init() {
}
