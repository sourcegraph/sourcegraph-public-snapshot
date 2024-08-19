// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RegistryNoCodeModules = (*registryNoCodeModules)(nil)

// RegistryNoCodeModules describes all the registry no-code module related methods that the Terraform
// Enterprise API supports.
//
// TFE API docs: (TODO: Add link to API docs)
type RegistryNoCodeModules interface {

	// Create a registry no-code module
	// **Note: This API is still in BETA and subject to change.**
	Create(ctx context.Context, organization string, options RegistryNoCodeModuleCreateOptions) (*RegistryNoCodeModule, error)

	// Read a registry no-code  module
	// **Note: This API is still in BETA and subject to change.**
	Read(ctx context.Context, noCodeModuleID string, options *RegistryNoCodeModuleReadOptions) (*RegistryNoCodeModule, error)

	// Update a registry no-code module
	// **Note: This API is still in BETA and subject to change.**
	Update(ctx context.Context, noCodeModuleID string, options RegistryNoCodeModuleUpdateOptions) (*RegistryNoCodeModule, error)

	// Delete a registry no-code module
	// **Note: This API is still in BETA and subject to change.**
	Delete(ctx context.Context, ID string) error
}

// registryNoCodeModules implements RegistryNoCodeModules.
type registryNoCodeModules struct {
	client *Client
}

// RegistryNoCodeModule represents a registry no-code module
type RegistryNoCodeModule struct {
	ID         string `jsonapi:"primary,no-code-modules"`
	VersionPin string `jsonapi:"attr,version-pin"`
	Enabled    bool   `jsonapi:"attr,enabled"`

	// Relations
	Organization    *Organization           `jsonapi:"relation,organization"`
	RegistryModule  *RegistryModule         `jsonapi:"relation,registry-module"`
	VariableOptions []*NoCodeVariableOption `jsonapi:"relation,variable-options"`
}

// NoCodeVariableOption represents a registry no-code module variable and its
// options.
type NoCodeVariableOption struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	Type string `jsonapi:"primary,variable-options"`

	// Required: The variable name
	VariableName string `jsonapi:"attr,variable-name"`

	// Required: The variable type
	VariableType string `jsonapi:"attr,variable-type"`

	// Optional: The options for the variable
	Options []string `jsonapi:"attr,options"`
}

// RegistryNoCodeModuleCreateOptions is used when creating a registry no-code module
type RegistryNoCodeModuleCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,no-code-modules"`

	// Required: the registry module to use for the no-code module (only the ID is used)
	RegistryModule *RegistryModule `jsonapi:"relation,registry-module"`

	// Optional: whether no-code is enabled for the module
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`

	// Optional: the version pin for the module. valid values are "latest" or a semver string
	VersionPin string `jsonapi:"attr,version-pin,omitempty"`

	// Optional: the variable options for the registry module
	VariableOptions []*NoCodeVariableOption `jsonapi:"relation,variable-options,omitempty"`
}

// RegistryNoCodeModuleIncludeOpt represents the available options for include query params.
type RegistryNoCodeModuleIncludeOpt string

var (
	// RegistryNoCodeIncludeVariableOptions is used to include variable options in the response
	RegistryNoCodeIncludeVariableOptions RegistryNoCodeModuleIncludeOpt = "variable-options"
)

// RegistryNoCodeModuleReadOptions is used when reading a registry no-code module
type RegistryNoCodeModuleReadOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-updating
	Type string `jsonapi:"primary,no-code-modules"`

	// Optional: Include is used to specify the related resources to include in the response.
	Include []RegistryNoCodeModuleIncludeOpt `url:"include,omitempty"`
}

// RegistryNoCodeModuleUpdateOptions is used when updating a registry no-code module
type RegistryNoCodeModuleUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-updating
	Type string `jsonapi:"primary,no-code-modules"`

	// Required: the registry module to use for the no-code module (only the ID is used)
	RegistryModule *RegistryModule `jsonapi:"relation,registry-module"`

	// Optional: the version pin for the module. valid values are "latest" or a semver string
	VersionPin string `jsonapi:"attr,version-pin,omitempty"`

	// Optional: whether no-code is enabled for the module
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`

	// Optional: are the variable options for the module
	VariableOptions []*NoCodeVariableOption `jsonapi:"relation,variable-options,omitempty"`
}

// Create a new registry no-code module
func (r *registryNoCodeModules) Create(ctx context.Context, organization string, options RegistryNoCodeModuleCreateOptions) (*RegistryNoCodeModule, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/no-code-modules", url.QueryEscape(organization))
	req, err := r.client.NewRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	rm := &RegistryNoCodeModule{}
	err = req.Do(ctx, rm)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

// Read a registry no-code module
func (r *registryNoCodeModules) Read(ctx context.Context, noCodeModuleID string, options *RegistryNoCodeModuleReadOptions) (*RegistryNoCodeModule, error) {
	if !validStringID(&noCodeModuleID) {
		return nil, ErrInvalidModuleID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("no-code-modules/%s", url.QueryEscape(noCodeModuleID))
	req, err := r.client.NewRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	rm := &RegistryNoCodeModule{}
	err = req.Do(ctx, rm)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

// Update a registry no-code module
func (r *registryNoCodeModules) Update(ctx context.Context, noCodeModuleID string, options RegistryNoCodeModuleUpdateOptions) (*RegistryNoCodeModule, error) {
	if !validString(&noCodeModuleID) {
		return nil, ErrInvalidModuleID
	}
	if !validStringID(&noCodeModuleID) {
		return nil, ErrInvalidModuleID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("no-code-modules/%s", url.QueryEscape(noCodeModuleID))
	req, err := r.client.NewRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	rm := &RegistryNoCodeModule{}
	err = req.Do(ctx, rm)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

// Delete is used to delete the registry no-code module
func (r *registryNoCodeModules) Delete(ctx context.Context, noCodeModuleID string) error {
	if !validStringID(&noCodeModuleID) {
		return ErrInvalidModuleID
	}

	u := fmt.Sprintf("no-code-modules/%s", url.QueryEscape(noCodeModuleID))
	req, err := r.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return req.Do(ctx, nil)
}

func (o RegistryNoCodeModuleCreateOptions) valid() error {
	if o.RegistryModule == nil || o.RegistryModule.ID == "" {
		return ErrRequiredRegistryModule
	}

	return nil
}

func (o *RegistryNoCodeModuleUpdateOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if o.RegistryModule == nil || o.RegistryModule.ID == "" {
		return ErrRequiredRegistryModule
	}

	return nil
}

func (o *RegistryNoCodeModuleReadOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validateNoCodeIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validateNoCodeIncludeParams(params []RegistryNoCodeModuleIncludeOpt) error {
	for _, p := range params {
		switch p {
		case RegistryNoCodeIncludeVariableOptions:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}
