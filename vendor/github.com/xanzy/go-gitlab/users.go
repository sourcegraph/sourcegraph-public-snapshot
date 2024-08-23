//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// List a couple of standard errors.
var (
	ErrUserActivatePrevented         = errors.New("Cannot activate a user that is blocked by admin or by LDAP synchronization")
	ErrUserApprovePrevented          = errors.New("Cannot approve a user that is blocked by admin or by LDAP synchronization")
	ErrUserBlockPrevented            = errors.New("Cannot block a user that is already blocked by LDAP synchronization")
	ErrUserConflict                  = errors.New("User does not have a pending request")
	ErrUserDeactivatePrevented       = errors.New("Cannot deactivate a user that is blocked by admin or by LDAP synchronization")
	ErrUserDisableTwoFactorPrevented = errors.New("Cannot disable two factor authentication if not authenticated as administrator")
	ErrUserNotFound                  = errors.New("User does not exist")
	ErrUserRejectPrevented           = errors.New("Cannot reject a user if not authenticated as administrator")
	ErrUserTwoFactorNotEnabled       = errors.New("Cannot disable two factor authentication if not enabled")
	ErrUserUnblockPrevented          = errors.New("Cannot unblock a user that is blocked by LDAP synchronization")
)

// UsersService handles communication with the user related methods of
// the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html
type UsersService struct {
	client *Client
}

// BasicUser included in other service responses (such as merge requests, pipelines, etc).
type BasicUser struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Name      string     `json:"name"`
	State     string     `json:"state"`
	CreatedAt *time.Time `json:"created_at"`
	AvatarURL string     `json:"avatar_url"`
	WebURL    string     `json:"web_url"`
}

// User represents a GitLab user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html
type User struct {
	ID                             int                `json:"id"`
	Username                       string             `json:"username"`
	Email                          string             `json:"email"`
	Name                           string             `json:"name"`
	State                          string             `json:"state"`
	WebURL                         string             `json:"web_url"`
	CreatedAt                      *time.Time         `json:"created_at"`
	Bio                            string             `json:"bio"`
	Bot                            bool               `json:"bot"`
	Location                       string             `json:"location"`
	PublicEmail                    string             `json:"public_email"`
	Skype                          string             `json:"skype"`
	Linkedin                       string             `json:"linkedin"`
	Twitter                        string             `json:"twitter"`
	WebsiteURL                     string             `json:"website_url"`
	Organization                   string             `json:"organization"`
	JobTitle                       string             `json:"job_title"`
	ExternUID                      string             `json:"extern_uid"`
	Provider                       string             `json:"provider"`
	ThemeID                        int                `json:"theme_id"`
	LastActivityOn                 *ISOTime           `json:"last_activity_on"`
	ColorSchemeID                  int                `json:"color_scheme_id"`
	IsAdmin                        bool               `json:"is_admin"`
	AvatarURL                      string             `json:"avatar_url"`
	CanCreateGroup                 bool               `json:"can_create_group"`
	CanCreateProject               bool               `json:"can_create_project"`
	ProjectsLimit                  int                `json:"projects_limit"`
	CurrentSignInAt                *time.Time         `json:"current_sign_in_at"`
	CurrentSignInIP                *net.IP            `json:"current_sign_in_ip"`
	LastSignInAt                   *time.Time         `json:"last_sign_in_at"`
	LastSignInIP                   *net.IP            `json:"last_sign_in_ip"`
	ConfirmedAt                    *time.Time         `json:"confirmed_at"`
	TwoFactorEnabled               bool               `json:"two_factor_enabled"`
	Note                           string             `json:"note"`
	Identities                     []*UserIdentity    `json:"identities"`
	External                       bool               `json:"external"`
	PrivateProfile                 bool               `json:"private_profile"`
	SharedRunnersMinutesLimit      int                `json:"shared_runners_minutes_limit"`
	ExtraSharedRunnersMinutesLimit int                `json:"extra_shared_runners_minutes_limit"`
	UsingLicenseSeat               bool               `json:"using_license_seat"`
	CustomAttributes               []*CustomAttribute `json:"custom_attributes"`
	NamespaceID                    int                `json:"namespace_id"`
}

// UserIdentity represents a user identity.
type UserIdentity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"extern_uid"`
}

// ListUsersOptions represents the available ListUsers() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-users
type ListUsersOptions struct {
	ListOptions
	Active          *bool `url:"active,omitempty" json:"active,omitempty"`
	Blocked         *bool `url:"blocked,omitempty" json:"blocked,omitempty"`
	ExcludeInternal *bool `url:"exclude_internal,omitempty" json:"exclude_internal,omitempty"`
	ExcludeExternal *bool `url:"exclude_external,omitempty" json:"exclude_external,omitempty"`

	// The options below are only available for admins.
	Search               *string    `url:"search,omitempty" json:"search,omitempty"`
	Username             *string    `url:"username,omitempty" json:"username,omitempty"`
	ExternalUID          *string    `url:"extern_uid,omitempty" json:"extern_uid,omitempty"`
	Provider             *string    `url:"provider,omitempty" json:"provider,omitempty"`
	CreatedBefore        *time.Time `url:"created_before,omitempty" json:"created_before,omitempty"`
	CreatedAfter         *time.Time `url:"created_after,omitempty" json:"created_after,omitempty"`
	OrderBy              *string    `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort                 *string    `url:"sort,omitempty" json:"sort,omitempty"`
	TwoFactor            *string    `url:"two_factor,omitempty" json:"two_factor,omitempty"`
	Admins               *bool      `url:"admins,omitempty" json:"admins,omitempty"`
	External             *bool      `url:"external,omitempty" json:"external,omitempty"`
	WithoutProjects      *bool      `url:"without_projects,omitempty" json:"without_projects,omitempty"`
	WithCustomAttributes *bool      `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
	WithoutProjectBots   *bool      `url:"without_project_bots,omitempty" json:"without_project_bots,omitempty"`
}

// ListUsers gets a list of users.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-users
func (s *UsersService) ListUsers(opt *ListUsersOptions, options ...RequestOptionFunc) ([]*User, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "users", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var usr []*User
	resp, err := s.client.Do(req, &usr)
	if err != nil {
		return nil, resp, err
	}

	return usr, resp, nil
}

// GetUsersOptions represents the available GetUser() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#single-user
type GetUsersOptions struct {
	WithCustomAttributes *bool `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
}

// GetUser gets a single user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#single-user
func (s *UsersService) GetUser(user int, opt GetUsersOptions, options ...RequestOptionFunc) (*User, *Response, error) {
	u := fmt.Sprintf("users/%d", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	usr := new(User)
	resp, err := s.client.Do(req, usr)
	if err != nil {
		return nil, resp, err
	}

	return usr, resp, nil
}

// CreateUserOptions represents the available CreateUser() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#user-creation
type CreateUserOptions struct {
	Email               *string `url:"email,omitempty" json:"email,omitempty"`
	Password            *string `url:"password,omitempty" json:"password,omitempty"`
	ResetPassword       *bool   `url:"reset_password,omitempty" json:"reset_password,omitempty"`
	ForceRandomPassword *bool   `url:"force_random_password,omitempty" json:"force_random_password,omitempty"`
	Username            *string `url:"username,omitempty" json:"username,omitempty"`
	Name                *string `url:"name,omitempty" json:"name,omitempty"`
	Skype               *string `url:"skype,omitempty" json:"skype,omitempty"`
	Linkedin            *string `url:"linkedin,omitempty" json:"linkedin,omitempty"`
	Twitter             *string `url:"twitter,omitempty" json:"twitter,omitempty"`
	WebsiteURL          *string `url:"website_url,omitempty" json:"website_url,omitempty"`
	Organization        *string `url:"organization,omitempty" json:"organization,omitempty"`
	JobTitle            *string `url:"job_title,omitempty" json:"job_title,omitempty"`
	ProjectsLimit       *int    `url:"projects_limit,omitempty" json:"projects_limit,omitempty"`
	ExternUID           *string `url:"extern_uid,omitempty" json:"extern_uid,omitempty"`
	Provider            *string `url:"provider,omitempty" json:"provider,omitempty"`
	Bio                 *string `url:"bio,omitempty" json:"bio,omitempty"`
	Location            *string `url:"location,omitempty" json:"location,omitempty"`
	Admin               *bool   `url:"admin,omitempty" json:"admin,omitempty"`
	CanCreateGroup      *bool   `url:"can_create_group,omitempty" json:"can_create_group,omitempty"`
	SkipConfirmation    *bool   `url:"skip_confirmation,omitempty" json:"skip_confirmation,omitempty"`
	External            *bool   `url:"external,omitempty" json:"external,omitempty"`
	PrivateProfile      *bool   `url:"private_profile,omitempty" json:"private_profile,omitempty"`
	Note                *string `url:"note,omitempty" json:"note,omitempty"`
	ThemeID             *int    `url:"theme_id,omitempty" json:"theme_id,omitempty"`
}

// CreateUser creates a new user. Note only administrators can create new users.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#user-creation
func (s *UsersService) CreateUser(opt *CreateUserOptions, options ...RequestOptionFunc) (*User, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "users", opt, options)
	if err != nil {
		return nil, nil, err
	}

	usr := new(User)
	resp, err := s.client.Do(req, usr)
	if err != nil {
		return nil, resp, err
	}

	return usr, resp, nil
}

// ModifyUserOptions represents the available ModifyUser() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#user-modification
type ModifyUserOptions struct {
	Email              *string `url:"email,omitempty" json:"email,omitempty"`
	Password           *string `url:"password,omitempty" json:"password,omitempty"`
	Username           *string `url:"username,omitempty" json:"username,omitempty"`
	Name               *string `url:"name,omitempty" json:"name,omitempty"`
	Skype              *string `url:"skype,omitempty" json:"skype,omitempty"`
	Linkedin           *string `url:"linkedin,omitempty" json:"linkedin,omitempty"`
	Twitter            *string `url:"twitter,omitempty" json:"twitter,omitempty"`
	WebsiteURL         *string `url:"website_url,omitempty" json:"website_url,omitempty"`
	Organization       *string `url:"organization,omitempty" json:"organization,omitempty"`
	JobTitle           *string `url:"job_title,omitempty" json:"job_title,omitempty"`
	ProjectsLimit      *int    `url:"projects_limit,omitempty" json:"projects_limit,omitempty"`
	ExternUID          *string `url:"extern_uid,omitempty" json:"extern_uid,omitempty"`
	Provider           *string `url:"provider,omitempty" json:"provider,omitempty"`
	Bio                *string `url:"bio,omitempty" json:"bio,omitempty"`
	Location           *string `url:"location,omitempty" json:"location,omitempty"`
	Admin              *bool   `url:"admin,omitempty" json:"admin,omitempty"`
	CanCreateGroup     *bool   `url:"can_create_group,omitempty" json:"can_create_group,omitempty"`
	SkipReconfirmation *bool   `url:"skip_reconfirmation,omitempty" json:"skip_reconfirmation,omitempty"`
	External           *bool   `url:"external,omitempty" json:"external,omitempty"`
	PrivateProfile     *bool   `url:"private_profile,omitempty" json:"private_profile,omitempty"`
	Note               *string `url:"note,omitempty" json:"note,omitempty"`
	ThemeID            *int    `url:"theme_id,omitempty" json:"theme_id,omitempty"`
	PublicEmail        *string `url:"public_email,omitempty" json:"public_email,omitempty"`
	CommitEmail        *string `url:"commit_email,omitempty" json:"commit_email,omitempty"`
}

// ModifyUser modifies an existing user. Only administrators can change attributes
// of a user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#user-modification
func (s *UsersService) ModifyUser(user int, opt *ModifyUserOptions, options ...RequestOptionFunc) (*User, *Response, error) {
	u := fmt.Sprintf("users/%d", user)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	usr := new(User)
	resp, err := s.client.Do(req, usr)
	if err != nil {
		return nil, resp, err
	}

	return usr, resp, nil
}

// DeleteUser deletes a user. Available only for administrators. This is an
// idempotent function, calling this function for a non-existent user id still
// returns a status code 200 OK. The JSON response differs if the user was
// actually deleted or not. In the former the user is returned and in the
// latter not.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#user-deletion
func (s *UsersService) DeleteUser(user int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("users/%d", user)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// CurrentUser gets currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-current-user
func (s *UsersService) CurrentUser(options ...RequestOptionFunc) (*User, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user", nil, options)
	if err != nil {
		return nil, nil, err
	}

	usr := new(User)
	resp, err := s.client.Do(req, usr)
	if err != nil {
		return nil, resp, err
	}

	return usr, resp, nil
}

// UserStatus represents the current status of a user
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#user-status
type UserStatus struct {
	Emoji        string            `json:"emoji"`
	Availability AvailabilityValue `json:"availability"`
	Message      string            `json:"message"`
	MessageHTML  string            `json:"message_html"`
}

// CurrentUserStatus retrieves the user status
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#user-status
func (s *UsersService) CurrentUserStatus(options ...RequestOptionFunc) (*UserStatus, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user/status", nil, options)
	if err != nil {
		return nil, nil, err
	}

	status := new(UserStatus)
	resp, err := s.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, nil
}

// GetUserStatus retrieves a user's status
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-the-status-of-a-user
func (s *UsersService) GetUserStatus(user int, options ...RequestOptionFunc) (*UserStatus, *Response, error) {
	u := fmt.Sprintf("users/%d/status", user)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	status := new(UserStatus)
	resp, err := s.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, nil
}

// UserStatusOptions represents the options required to set the status
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#set-user-status
type UserStatusOptions struct {
	Emoji        *string            `url:"emoji,omitempty" json:"emoji,omitempty"`
	Availability *AvailabilityValue `url:"availability,omitempty" json:"availability,omitempty"`
	Message      *string            `url:"message,omitempty" json:"message,omitempty"`
}

// SetUserStatus sets the user's status
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#set-user-status
func (s *UsersService) SetUserStatus(opt *UserStatusOptions, options ...RequestOptionFunc) (*UserStatus, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPut, "user/status", opt, options)
	if err != nil {
		return nil, nil, err
	}

	status := new(UserStatus)
	resp, err := s.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, nil
}

// SSHKey represents a SSH key.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-ssh-keys
type SSHKey struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Key       string     `json:"key"`
	CreatedAt *time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// ListSSHKeys gets a list of currently authenticated user's SSH keys.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-ssh-keys
func (s *UsersService) ListSSHKeys(options ...RequestOptionFunc) ([]*SSHKey, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user/keys", nil, options)
	if err != nil {
		return nil, nil, err
	}

	var k []*SSHKey
	resp, err := s.client.Do(req, &k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// ListSSHKeysForUserOptions represents the available ListSSHKeysForUser() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#list-ssh-keys-for-user
type ListSSHKeysForUserOptions ListOptions

// ListSSHKeysForUser gets a list of a specified user's SSH keys.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#list-ssh-keys-for-user
func (s *UsersService) ListSSHKeysForUser(uid interface{}, opt *ListSSHKeysForUserOptions, options ...RequestOptionFunc) ([]*SSHKey, *Response, error) {
	user, err := parseID(uid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("users/%s/keys", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var k []*SSHKey
	resp, err := s.client.Do(req, &k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// GetSSHKey gets a single key.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#single-ssh-key
func (s *UsersService) GetSSHKey(key int, options ...RequestOptionFunc) (*SSHKey, *Response, error) {
	u := fmt.Sprintf("user/keys/%d", key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(SSHKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// GetSSHKeyForUser gets a single key for a given user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#single-ssh-key-for-given-user
func (s *UsersService) GetSSHKeyForUser(user int, key int, options ...RequestOptionFunc) (*SSHKey, *Response, error) {
	u := fmt.Sprintf("users/%d/keys/%d", user, key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(SSHKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// AddSSHKeyOptions represents the available AddSSHKey() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-ssh-key
type AddSSHKeyOptions struct {
	Title     *string  `url:"title,omitempty" json:"title,omitempty"`
	Key       *string  `url:"key,omitempty" json:"key,omitempty"`
	ExpiresAt *ISOTime `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// AddSSHKey creates a new key owned by the currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-ssh-key
func (s *UsersService) AddSSHKey(opt *AddSSHKeyOptions, options ...RequestOptionFunc) (*SSHKey, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "user/keys", opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(SSHKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// AddSSHKeyForUser creates new key owned by specified user. Available only for
// admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-ssh-key-for-user
func (s *UsersService) AddSSHKeyForUser(user int, opt *AddSSHKeyOptions, options ...RequestOptionFunc) (*SSHKey, *Response, error) {
	u := fmt.Sprintf("users/%d/keys", user)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(SSHKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// DeleteSSHKey deletes key owned by currently authenticated user. This is an
// idempotent function and calling it on a key that is already deleted or not
// available results in 200 OK.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#delete-ssh-key-for-current-user
func (s *UsersService) DeleteSSHKey(key int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("user/keys/%d", key)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteSSHKeyForUser deletes key owned by a specified user. Available only
// for admin.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#delete-ssh-key-for-given-user
func (s *UsersService) DeleteSSHKeyForUser(user, key int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("users/%d/keys/%d", user, key)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// GPGKey represents a GPG key.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-all-gpg-keys
type GPGKey struct {
	ID        int        `json:"id"`
	Key       string     `json:"key"`
	CreatedAt *time.Time `json:"created_at"`
}

// ListGPGKeys gets a list of currently authenticated user’s GPG keys.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-all-gpg-keys
func (s *UsersService) ListGPGKeys(options ...RequestOptionFunc) ([]*GPGKey, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user/gpg_keys", nil, options)
	if err != nil {
		return nil, nil, err
	}

	var ks []*GPGKey
	resp, err := s.client.Do(req, &ks)
	if err != nil {
		return nil, resp, err
	}

	return ks, resp, nil
}

// GetGPGKey gets a specific GPG key of currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#get-a-specific-gpg-key
func (s *UsersService) GetGPGKey(key int, options ...RequestOptionFunc) (*GPGKey, *Response, error) {
	u := fmt.Sprintf("user/gpg_keys/%d", key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(GPGKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// AddGPGKeyOptions represents the available AddGPGKey() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-a-gpg-key
type AddGPGKeyOptions struct {
	Key *string `url:"key,omitempty" json:"key,omitempty"`
}

// AddGPGKey creates a new GPG key owned by the currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-a-gpg-key
func (s *UsersService) AddGPGKey(opt *AddGPGKeyOptions, options ...RequestOptionFunc) (*GPGKey, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "user/gpg_keys", opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(GPGKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// DeleteGPGKey deletes a GPG key owned by currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#delete-a-gpg-key
func (s *UsersService) DeleteGPGKey(key int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("user/gpg_keys/%d", key)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListGPGKeysForUser gets a list of a specified user’s GPG keys.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#list-all-gpg-keys-for-given-user
func (s *UsersService) ListGPGKeysForUser(user int, options ...RequestOptionFunc) ([]*GPGKey, *Response, error) {
	u := fmt.Sprintf("users/%d/gpg_keys", user)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var ks []*GPGKey
	resp, err := s.client.Do(req, &ks)
	if err != nil {
		return nil, resp, err
	}

	return ks, resp, nil
}

// GetGPGKeyForUser gets a specific GPG key for a given user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#get-a-specific-gpg-key-for-a-given-user
func (s *UsersService) GetGPGKeyForUser(user, key int, options ...RequestOptionFunc) (*GPGKey, *Response, error) {
	u := fmt.Sprintf("users/%d/gpg_keys/%d", user, key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(GPGKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// AddGPGKeyForUser creates new GPG key owned by the specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#add-a-gpg-key-for-a-given-user
func (s *UsersService) AddGPGKeyForUser(user int, opt *AddGPGKeyOptions, options ...RequestOptionFunc) (*GPGKey, *Response, error) {
	u := fmt.Sprintf("users/%d/gpg_keys", user)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(GPGKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// DeleteGPGKeyForUser deletes a GPG key owned by a specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#delete-a-gpg-key-for-a-given-user
func (s *UsersService) DeleteGPGKeyForUser(user, key int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("users/%d/gpg_keys/%d", user, key)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// Email represents an Email.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-emails
type Email struct {
	ID          int        `json:"id"`
	Email       string     `json:"email"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
}

// ListEmails gets a list of currently authenticated user's Emails.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#list-emails
func (s *UsersService) ListEmails(options ...RequestOptionFunc) ([]*Email, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user/emails", nil, options)
	if err != nil {
		return nil, nil, err
	}

	var e []*Email
	resp, err := s.client.Do(req, &e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// ListEmailsForUserOptions represents the available ListEmailsForUser() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#list-emails-for-user
type ListEmailsForUserOptions ListOptions

// ListEmailsForUser gets a list of a specified user's Emails. Available
// only for admin
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#list-emails-for-user
func (s *UsersService) ListEmailsForUser(user int, opt *ListEmailsForUserOptions, options ...RequestOptionFunc) ([]*Email, *Response, error) {
	u := fmt.Sprintf("users/%d/emails", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var e []*Email
	resp, err := s.client.Do(req, &e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// GetEmail gets a single email.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#single-email
func (s *UsersService) GetEmail(email int, options ...RequestOptionFunc) (*Email, *Response, error) {
	u := fmt.Sprintf("user/emails/%d", email)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	e := new(Email)
	resp, err := s.client.Do(req, e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// AddEmailOptions represents the available AddEmail() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-email
type AddEmailOptions struct {
	Email            *string `url:"email,omitempty" json:"email,omitempty"`
	SkipConfirmation *bool   `url:"skip_confirmation,omitempty" json:"skip_confirmation,omitempty"`
}

// AddEmail creates a new email owned by the currently authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-email
func (s *UsersService) AddEmail(opt *AddEmailOptions, options ...RequestOptionFunc) (*Email, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "user/emails", opt, options)
	if err != nil {
		return nil, nil, err
	}

	e := new(Email)
	resp, err := s.client.Do(req, e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// AddEmailForUser creates new email owned by specified user. Available only for
// admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#add-email-for-user
func (s *UsersService) AddEmailForUser(user int, opt *AddEmailOptions, options ...RequestOptionFunc) (*Email, *Response, error) {
	u := fmt.Sprintf("users/%d/emails", user)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	e := new(Email)
	resp, err := s.client.Do(req, e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// DeleteEmail deletes email owned by currently authenticated user. This is an
// idempotent function and calling it on a key that is already deleted or not
// available results in 200 OK.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#delete-email-for-current-user
func (s *UsersService) DeleteEmail(email int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("user/emails/%d", email)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteEmailForUser deletes email owned by a specified user. Available only
// for admin.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#delete-email-for-given-user
func (s *UsersService) DeleteEmailForUser(user, email int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("users/%d/emails/%d", user, email)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// BlockUser blocks the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#block-user
func (s *UsersService) BlockUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/block", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 403:
		return ErrUserBlockPrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// UnblockUser unblocks the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#unblock-user
func (s *UsersService) UnblockUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/unblock", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 403:
		return ErrUserUnblockPrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// BanUser bans the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#ban-user
func (s *UsersService) BanUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/ban", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// UnbanUser unbans the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#unban-user
func (s *UsersService) UnbanUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/unban", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// DeactivateUser deactivate the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#deactivate-user
func (s *UsersService) DeactivateUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/deactivate", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 403:
		return ErrUserDeactivatePrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// ActivateUser activate the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#activate-user
func (s *UsersService) ActivateUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/activate", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 403:
		return ErrUserActivatePrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// ApproveUser approve the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#approve-user
func (s *UsersService) ApproveUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/approve", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		return nil
	case 403:
		return ErrUserApprovePrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// RejectUser reject the specified user. Available only for admin.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/users.html#reject-user
func (s *UsersService) RejectUser(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/reject", user)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		return nil
	case 403:
		return ErrUserRejectPrevented
	case 404:
		return ErrUserNotFound
	case 409:
		return ErrUserConflict
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}

// ImpersonationToken represents an impersonation token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-all-impersonation-tokens-of-a-user
type ImpersonationToken struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Active    bool       `json:"active"`
	Token     string     `json:"token"`
	Scopes    []string   `json:"scopes"`
	Revoked   bool       `json:"revoked"`
	CreatedAt *time.Time `json:"created_at"`
	ExpiresAt *ISOTime   `json:"expires_at"`
}

// GetAllImpersonationTokensOptions represents the available
// GetAllImpersonationTokens() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-all-impersonation-tokens-of-a-user
type GetAllImpersonationTokensOptions struct {
	ListOptions
	State *string `url:"state,omitempty" json:"state,omitempty"`
}

// GetAllImpersonationTokens retrieves all impersonation tokens of a user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-all-impersonation-tokens-of-a-user
func (s *UsersService) GetAllImpersonationTokens(user int, opt *GetAllImpersonationTokensOptions, options ...RequestOptionFunc) ([]*ImpersonationToken, *Response, error) {
	u := fmt.Sprintf("users/%d/impersonation_tokens", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ts []*ImpersonationToken
	resp, err := s.client.Do(req, &ts)
	if err != nil {
		return nil, resp, err
	}

	return ts, resp, nil
}

// GetImpersonationToken retrieves an impersonation token of a user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-an-impersonation-token-of-a-user
func (s *UsersService) GetImpersonationToken(user, token int, options ...RequestOptionFunc) (*ImpersonationToken, *Response, error) {
	u := fmt.Sprintf("users/%d/impersonation_tokens/%d", user, token)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(ImpersonationToken)
	resp, err := s.client.Do(req, &t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// CreateImpersonationTokenOptions represents the available
// CreateImpersonationToken() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-an-impersonation-token
type CreateImpersonationTokenOptions struct {
	Name      *string    `url:"name,omitempty" json:"name,omitempty"`
	Scopes    *[]string  `url:"scopes,omitempty" json:"scopes,omitempty"`
	ExpiresAt *time.Time `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// CreateImpersonationToken creates an impersonation token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-an-impersonation-token
func (s *UsersService) CreateImpersonationToken(user int, opt *CreateImpersonationTokenOptions, options ...RequestOptionFunc) (*ImpersonationToken, *Response, error) {
	u := fmt.Sprintf("users/%d/impersonation_tokens", user)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(ImpersonationToken)
	resp, err := s.client.Do(req, &t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// RevokeImpersonationToken revokes an impersonation token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#revoke-an-impersonation-token
func (s *UsersService) RevokeImpersonationToken(user, token int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("users/%d/impersonation_tokens/%d", user, token)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// CreatePersonalAccessTokenOptions represents the available
// CreatePersonalAccessToken() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-personal-access-token
type CreatePersonalAccessTokenOptions struct {
	Name      *string   `url:"name,omitempty" json:"name,omitempty"`
	ExpiresAt *ISOTime  `url:"expires_at,omitempty" json:"expires_at,omitempty"`
	Scopes    *[]string `url:"scopes,omitempty" json:"scopes,omitempty"`
}

// CreatePersonalAccessToken creates a personal access token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#create-a-personal-access-token
func (s *UsersService) CreatePersonalAccessToken(user int, opt *CreatePersonalAccessTokenOptions, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	u := fmt.Sprintf("users/%d/personal_access_tokens", user)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(PersonalAccessToken)
	resp, err := s.client.Do(req, &t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// UserActivity represents an entry in the user/activities response
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-user-activities
type UserActivity struct {
	Username       string   `json:"username"`
	LastActivityOn *ISOTime `json:"last_activity_on"`
}

// GetUserActivitiesOptions represents the options for GetUserActivities
//
// GitLap API docs:
// https://docs.gitlab.com/ee/api/users.html#get-user-activities
type GetUserActivitiesOptions struct {
	ListOptions
	From *ISOTime `url:"from,omitempty" json:"from,omitempty"`
}

// GetUserActivities retrieves user activities (admin only)
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#get-user-activities
func (s *UsersService) GetUserActivities(opt *GetUserActivitiesOptions, options ...RequestOptionFunc) ([]*UserActivity, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "user/activities", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var t []*UserActivity
	resp, err := s.client.Do(req, &t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// UserMembership represents a membership of the user in a namespace or project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#user-memberships
type UserMembership struct {
	SourceID    int              `json:"source_id"`
	SourceName  string           `json:"source_name"`
	SourceType  string           `json:"source_type"`
	AccessLevel AccessLevelValue `json:"access_level"`
}

// GetUserMembershipOptions represents the options available to query user memberships.
//
// GitLab API docs:
// ohttps://docs.gitlab.com/ee/api/users.html#user-memberships
type GetUserMembershipOptions struct {
	ListOptions
	Type *string `url:"type,omitempty" json:"type,omitempty"`
}

// GetUserMemberships retrieves a list of the user's memberships.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#user-memberships
func (s *UsersService) GetUserMemberships(user int, opt *GetUserMembershipOptions, options ...RequestOptionFunc) ([]*UserMembership, *Response, error) {
	u := fmt.Sprintf("users/%d/memberships", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var m []*UserMembership
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// DisableTwoFactor disables two factor authentication for the specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/users.html#disable-two-factor-authentication
func (s *UsersService) DisableTwoFactor(user int, options ...RequestOptionFunc) error {
	u := fmt.Sprintf("users/%d/disable_two_factor", user)

	req, err := s.client.NewRequest(http.MethodPatch, u, nil, options)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil && resp == nil {
		return err
	}

	switch resp.StatusCode {
	case 204:
		return nil
	case 400:
		return ErrUserTwoFactorNotEnabled
	case 403:
		return ErrUserDisableTwoFactorPrevented
	case 404:
		return ErrUserNotFound
	default:
		return fmt.Errorf("Received unexpected result code: %d", resp.StatusCode)
	}
}
