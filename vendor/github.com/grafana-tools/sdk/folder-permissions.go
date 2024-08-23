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

type PermissionType uint

const PermissionView = PermissionType(1)
const PermissionEdit = PermissionType(2)
const PermissionAdmin = PermissionType(4)

type FolderPermission struct {
	Id             uint           `json:"id"`
	FolderId       uint           `json:"folderId"`
	Created        string         `json:"created"`
	Updated        string         `json:"updated"`
	UserId         uint           `json:"userId,omitempty"`
	UserLogin      string         `json:"userLogin,omitempty"`
	UserEmail      string         `json:"userEmail,omitempty"`
	TeamId         uint           `json:"teamId,omitempty"`
	Team           string         `json:"team,omitempty"`
	Role           string         `json:"role,omitempty"`
	Permission     PermissionType `json:"permission"`
	PermissionName string         `json:"permissionName"`
	Uid            string         `json:"uid,omitempty"`
	Title          string         `json:"title,omitempty"`
	Slug           string         `json:"slug,omitempty"`
	IsFolder       bool           `json:"isFolder"`
	Url            string         `json:"url,omitempty"`
}
