package kernel

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// LoadProps holds the necessary information to load a library into the
// @jsii/kernel process through the Load method.
type LoadProps struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// LoadResponse contains the data returned by the @jsii/kernel process in
// response to a load request.
type LoadResponse struct {
	kernelResponse
	Assembly string  `json:"assembly"`
	Types    float64 `json:"types"`
}

// Load ensures the specified assembly has been loaded into the @jsii/kernel
// process. This call is idempotent (calling it several times with the same
// input results in the same output).
func (c *Client) Load(props LoadProps, tarball []byte) (response LoadResponse, err error) {
	if response, cached := c.loaded[props]; cached {
		return response, nil
	}

	tmpfile, err := ioutil.TempFile("", fmt.Sprintf(
		"%v-%v.*.tgz",
		regexp.MustCompile("[^a-zA-Z0-9_-]").ReplaceAllString(props.Name, "-"),
		version,
	))
	if err != nil {
		return
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(tarball); err != nil {
		panic(err)
	}
	tmpfile.Close()

	type request struct {
		kernelRequest
		LoadProps
		Tarball string `json:"tarball"`
	}
	err = c.request(request{kernelRequest{"load"}, props, tmpfile.Name()}, &response)

	if err == nil {
		c.loaded[props] = response
	}

	return
}

// UnmarshalJSON provides custom unmarshalling implementation for response
// structs. Creating new types is required in order to avoid infinite recursion.
func (r *LoadResponse) UnmarshalJSON(data []byte) error {
	type response LoadResponse
	return unmarshalKernelResponse(data, (*response)(r), r)
}
