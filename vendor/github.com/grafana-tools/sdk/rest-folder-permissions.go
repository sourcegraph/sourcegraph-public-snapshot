package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2022 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"context"
	"encoding/json"
	"fmt"
)

// https://grafana.com/docs/grafana/latest/http_api/folder_permissions/

// GetFolderPermissions gets permissions for a folder.
// Reflects GET /api/folders/:uid/permissions API call.
func (r *Client) GetFolderPermissions(ctx context.Context, folderUID string) ([]FolderPermission, error) {
	var (
		raw  []byte
		fs   []FolderPermission
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/folders/%s/permissions", folderUID), nil); err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &fs)
	return fs, err
}

// UpdateFolderPermissions update folders permission
// Reflects PUT /api/folders/:uid/permissions API call.
func (r *Client) UpdateFolderPermissions(ctx context.Context, folderUID string, up ...FolderPermission) (StatusMessage, error) {
	var (
		raw  []byte
		rf   StatusMessage
		code int
		err  error
	)
	request := struct {
		Items []FolderPermission `json:"items"`
	}{
		Items: up,
	}
	if raw, err = json.Marshal(request); err != nil {
		return rf, err
	}
	if raw, code, err = r.post(ctx, fmt.Sprintf("api/folders/%s/permissions", folderUID), nil, raw); err != nil {
		return rf, err
	}
	if code != 200 {
		return rf, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	err = json.Unmarshal(raw, &rf)
	return rf, err
}
