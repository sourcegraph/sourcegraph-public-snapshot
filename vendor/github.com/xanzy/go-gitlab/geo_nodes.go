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
	"fmt"
	"net/http"
)

// GeoNode represents a GitLab Geo Node.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/geo_nodes.html
type GeoNode struct {
	ID                               int          `json:"id"`
	Name                             string       `json:"name"`
	URL                              string       `json:"url"`
	InternalURL                      string       `json:"internal_url"`
	Primary                          bool         `json:"primary"`
	Enabled                          bool         `json:"enabled"`
	Current                          bool         `json:"current"`
	FilesMaxCapacity                 int          `json:"files_max_capacity"`
	ReposMaxCapacity                 int          `json:"repos_max_capacity"`
	VerificationMaxCapacity          int          `json:"verification_max_capacity"`
	SelectiveSyncType                string       `json:"selective_sync_type"`
	SelectiveSyncShards              []string     `json:"selective_sync_shards"`
	SelectiveSyncNamespaceIds        []int        `json:"selective_sync_namespace_ids"`
	MinimumReverificationInterval    int          `json:"minimum_reverification_interval"`
	ContainerRepositoriesMaxCapacity int          `json:"container_repositories_max_capacity"`
	SyncObjectStorage                bool         `json:"sync_object_storage"`
	CloneProtocol                    string       `json:"clone_protocol"`
	WebEditURL                       string       `json:"web_edit_url"`
	WebGeoProjectsURL                string       `json:"web_geo_projects_url"`
	Links                            GeoNodeLinks `json:"_links"`
}

// GeoNodeLinks represents links for GitLab GeoNode.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/geo_nodes.html
type GeoNodeLinks struct {
	Self   string `json:"self"`
	Status string `json:"status"`
	Repair string `json:"repair"`
}

// GeoNodesService handles communication with Geo Nodes related methods
// of GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/geo_nodes.html
type GeoNodesService struct {
	client *Client
}

// CreateGeoNodesOptions represents the available CreateGeoNode() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#create-a-new-geo-node
type CreateGeoNodesOptions struct {
	Primary                          *bool     `url:"primary,omitempty" json:"primary,omitempty"`
	Enabled                          *bool     `url:"enabled,omitempty" json:"enabled,omitempty"`
	Name                             *string   `url:"name,omitempty" json:"name,omitempty"`
	URL                              *string   `url:"url,omitempty" json:"url,omitempty"`
	InternalURL                      *string   `url:"internal_url,omitempty" json:"internal_url,omitempty"`
	FilesMaxCapacity                 *int      `url:"files_max_capacity,omitempty" json:"files_max_capacity,omitempty"`
	ReposMaxCapacity                 *int      `url:"repos_max_capacity,omitempty" json:"repos_max_capacity,omitempty"`
	VerificationMaxCapacity          *int      `url:"verification_max_capacity,omitempty" json:"verification_max_capacity,omitempty"`
	ContainerRepositoriesMaxCapacity *int      `url:"container_repositories_max_capacity,omitempty" json:"container_repositories_max_capacity,omitempty"`
	SyncObjectStorage                *bool     `url:"sync_object_storage,omitempty" json:"sync_object_storage,omitempty"`
	SelectiveSyncType                *string   `url:"selective_sync_type,omitempty" json:"selective_sync_type,omitempty"`
	SelectiveSyncShards              *[]string `url:"selective_sync_shards,omitempty" json:"selective_sync_shards,omitempty"`
	SelectiveSyncNamespaceIds        *[]int    `url:"selective_sync_namespace_ids,omitempty" json:"selective_sync_namespace_ids,omitempty"`
	MinimumReverificationInterval    *int      `url:"minimum_reverification_interval,omitempty" json:"minimum_reverification_interval,omitempty"`
}

// CreateGeoNode creates a new Geo Node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#create-a-new-geo-node
func (s *GeoNodesService) CreateGeoNode(opt *CreateGeoNodesOptions, options ...RequestOptionFunc) (*GeoNode, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "geo_nodes", opt, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(GeoNode)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// ListGeoNodesOptions represents the available ListGeoNodes() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-configuration-about-all-geo-nodes
type ListGeoNodesOptions ListOptions

// ListGeoNodes gets a list of geo nodes.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-configuration-about-all-geo-nodes
func (s *GeoNodesService) ListGeoNodes(opt *ListGeoNodesOptions, options ...RequestOptionFunc) ([]*GeoNode, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "geo_nodes", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*GeoNode
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// GetGeoNode gets a specific geo node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-configuration-about-a-specific-geo-node
func (s *GeoNodesService) GetGeoNode(id int, options ...RequestOptionFunc) (*GeoNode, *Response, error) {
	u := fmt.Sprintf("geo_nodes/%d", id)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(GeoNode)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// UpdateGeoNodesOptions represents the available EditGeoNode() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#edit-a-geo-node
type UpdateGeoNodesOptions struct {
	ID                               *int      `url:"primary,omitempty" json:"primary,omitempty"`
	Enabled                          *bool     `url:"enabled,omitempty" json:"enabled,omitempty"`
	Name                             *string   `url:"name,omitempty" json:"name,omitempty"`
	URL                              *string   `url:"url,omitempty" json:"url,omitempty"`
	InternalURL                      *string   `url:"internal_url,omitempty" json:"internal_url,omitempty"`
	FilesMaxCapacity                 *int      `url:"files_max_capacity,omitempty" json:"files_max_capacity,omitempty"`
	ReposMaxCapacity                 *int      `url:"repos_max_capacity,omitempty" json:"repos_max_capacity,omitempty"`
	VerificationMaxCapacity          *int      `url:"verification_max_capacity,omitempty" json:"verification_max_capacity,omitempty"`
	ContainerRepositoriesMaxCapacity *int      `url:"container_repositories_max_capacity,omitempty" json:"container_repositories_max_capacity,omitempty"`
	SyncObjectStorage                *bool     `url:"sync_object_storage,omitempty" json:"sync_object_storage,omitempty"`
	SelectiveSyncType                *string   `url:"selective_sync_type,omitempty" json:"selective_sync_type,omitempty"`
	SelectiveSyncShards              *[]string `url:"selective_sync_shards,omitempty" json:"selective_sync_shards,omitempty"`
	SelectiveSyncNamespaceIds        *[]int    `url:"selective_sync_namespace_ids,omitempty" json:"selective_sync_namespace_ids,omitempty"`
	MinimumReverificationInterval    *int      `url:"minimum_reverification_interval,omitempty" json:"minimum_reverification_interval,omitempty"`
}

// EditGeoNode updates settings of an existing Geo node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#edit-a-geo-node
func (s *GeoNodesService) EditGeoNode(id int, opt *UpdateGeoNodesOptions, options ...RequestOptionFunc) (*GeoNode, *Response, error) {
	u := fmt.Sprintf("geo_nodes/%d", id)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(GeoNode)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// DeleteGeoNode removes the Geo node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#delete-a-geo-node
func (s *GeoNodesService) DeleteGeoNode(id int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("geo_nodes/%d", id)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// RepairGeoNode to repair the OAuth authentication of a Geo node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#repair-a-geo-node
func (s *GeoNodesService) RepairGeoNode(id int, options ...RequestOptionFunc) (*GeoNode, *Response, error) {
	u := fmt.Sprintf("geo_nodes/%d/repair", id)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(GeoNode)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// GeoNodeStatus represents the status of Geo Node.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-status-about-all-geo-nodes
type GeoNodeStatus struct {
	GeoNodeID                                     int    `json:"geo_node_id"`
	Healthy                                       bool   `json:"healthy"`
	Health                                        string `json:"health"`
	HealthStatus                                  string `json:"health_status"`
	MissingOauthApplication                       bool   `json:"missing_oauth_application"`
	AttachmentsCount                              int    `json:"attachments_count"`
	AttachmentsSyncedCount                        int    `json:"attachments_synced_count"`
	AttachmentsFailedCount                        int    `json:"attachments_failed_count"`
	AttachmentsSyncedMissingOnPrimaryCount        int    `json:"attachments_synced_missing_on_primary_count"`
	AttachmentsSyncedInPercentage                 string `json:"attachments_synced_in_percentage"`
	DbReplicationLagSeconds                       int    `json:"db_replication_lag_seconds"`
	LfsObjectsCount                               int    `json:"lfs_objects_count"`
	LfsObjectsSyncedCount                         int    `json:"lfs_objects_synced_count"`
	LfsObjectsFailedCount                         int    `json:"lfs_objects_failed_count"`
	LfsObjectsSyncedMissingOnPrimaryCount         int    `json:"lfs_objects_synced_missing_on_primary_count"`
	LfsObjectsSyncedInPercentage                  string `json:"lfs_objects_synced_in_percentage"`
	JobArtifactsCount                             int    `json:"job_artifacts_count"`
	JobArtifactsSyncedCount                       int    `json:"job_artifacts_synced_count"`
	JobArtifactsFailedCount                       int    `json:"job_artifacts_failed_count"`
	JobArtifactsSyncedMissingOnPrimaryCount       int    `json:"job_artifacts_synced_missing_on_primary_count"`
	JobArtifactsSyncedInPercentage                string `json:"job_artifacts_synced_in_percentage"`
	ContainerRepositoriesCount                    int    `json:"container_repositories_count"`
	ContainerRepositoriesSyncedCount              int    `json:"container_repositories_synced_count"`
	ContainerRepositoriesFailedCount              int    `json:"container_repositories_failed_count"`
	ContainerRepositoriesSyncedInPercentage       string `json:"container_repositories_synced_in_percentage"`
	DesignRepositoriesCount                       int    `json:"design_repositories_count"`
	DesignRepositoriesSyncedCount                 int    `json:"design_repositories_synced_count"`
	DesignRepositoriesFailedCount                 int    `json:"design_repositories_failed_count"`
	DesignRepositoriesSyncedInPercentage          string `json:"design_repositories_synced_in_percentage"`
	ProjectsCount                                 int    `json:"projects_count"`
	RepositoriesCount                             int    `json:"repositories_count"`
	RepositoriesFailedCount                       int    `json:"repositories_failed_count"`
	RepositoriesSyncedCount                       int    `json:"repositories_synced_count"`
	RepositoriesSyncedInPercentage                string `json:"repositories_synced_in_percentage"`
	WikisCount                                    int    `json:"wikis_count"`
	WikisFailedCount                              int    `json:"wikis_failed_count"`
	WikisSyncedCount                              int    `json:"wikis_synced_count"`
	WikisSyncedInPercentage                       string `json:"wikis_synced_in_percentage"`
	ReplicationSlotsCount                         int    `json:"replication_slots_count"`
	ReplicationSlotsUsedCount                     int    `json:"replication_slots_used_count"`
	ReplicationSlotsUsedInPercentage              string `json:"replication_slots_used_in_percentage"`
	ReplicationSlotsMaxRetainedWalBytes           int    `json:"replication_slots_max_retained_wal_bytes"`
	RepositoriesCheckedCount                      int    `json:"repositories_checked_count"`
	RepositoriesCheckedFailedCount                int    `json:"repositories_checked_failed_count"`
	RepositoriesCheckedInPercentage               string `json:"repositories_checked_in_percentage"`
	RepositoriesChecksummedCount                  int    `json:"repositories_checksummed_count"`
	RepositoriesChecksumFailedCount               int    `json:"repositories_checksum_failed_count"`
	RepositoriesChecksummedInPercentage           string `json:"repositories_checksummed_in_percentage"`
	WikisChecksummedCount                         int    `json:"wikis_checksummed_count"`
	WikisChecksumFailedCount                      int    `json:"wikis_checksum_failed_count"`
	WikisChecksummedInPercentage                  string `json:"wikis_checksummed_in_percentage"`
	RepositoriesVerifiedCount                     int    `json:"repositories_verified_count"`
	RepositoriesVerificationFailedCount           int    `json:"repositories_verification_failed_count"`
	RepositoriesVerifiedInPercentage              string `json:"repositories_verified_in_percentage"`
	RepositoriesChecksumMismatchCount             int    `json:"repositories_checksum_mismatch_count"`
	WikisVerifiedCount                            int    `json:"wikis_verified_count"`
	WikisVerificationFailedCount                  int    `json:"wikis_verification_failed_count"`
	WikisVerifiedInPercentage                     string `json:"wikis_verified_in_percentage"`
	WikisChecksumMismatchCount                    int    `json:"wikis_checksum_mismatch_count"`
	RepositoriesRetryingVerificationCount         int    `json:"repositories_retrying_verification_count"`
	WikisRetryingVerificationCount                int    `json:"wikis_retrying_verification_count"`
	LastEventID                                   int    `json:"last_event_id"`
	LastEventTimestamp                            int    `json:"last_event_timestamp"`
	CursorLastEventID                             int    `json:"cursor_last_event_id"`
	CursorLastEventTimestamp                      int    `json:"cursor_last_event_timestamp"`
	LastSuccessfulStatusCheckTimestamp            int    `json:"last_successful_status_check_timestamp"`
	Version                                       string `json:"version"`
	Revision                                      string `json:"revision"`
	MergeRequestDiffsCount                        int    `json:"merge_request_diffs_count"`
	MergeRequestDiffsChecksumTotalCount           int    `json:"merge_request_diffs_checksum_total_count"`
	MergeRequestDiffsChecksummedCount             int    `json:"merge_request_diffs_checksummed_count"`
	MergeRequestDiffsChecksumFailedCount          int    `json:"merge_request_diffs_checksum_failed_count"`
	MergeRequestDiffsSyncedCount                  int    `json:"merge_request_diffs_synced_count"`
	MergeRequestDiffsFailedCount                  int    `json:"merge_request_diffs_failed_count"`
	MergeRequestDiffsRegistryCount                int    `json:"merge_request_diffs_registry_count"`
	MergeRequestDiffsVerificationTotalCount       int    `json:"merge_request_diffs_verification_total_count"`
	MergeRequestDiffsVerifiedCount                int    `json:"merge_request_diffs_verified_count"`
	MergeRequestDiffsVerificationFailedCount      int    `json:"merge_request_diffs_verification_failed_count"`
	MergeRequestDiffsSyncedInPercentage           string `json:"merge_request_diffs_synced_in_percentage"`
	MergeRequestDiffsVerifiedInPercentage         string `json:"merge_request_diffs_verified_in_percentage"`
	PackageFilesCount                             int    `json:"package_files_count"`
	PackageFilesChecksumTotalCount                int    `json:"package_files_checksum_total_count"`
	PackageFilesChecksummedCount                  int    `json:"package_files_checksummed_count"`
	PackageFilesChecksumFailedCount               int    `json:"package_files_checksum_failed_count"`
	PackageFilesSyncedCount                       int    `json:"package_files_synced_count"`
	PackageFilesFailedCount                       int    `json:"package_files_failed_count"`
	PackageFilesRegistryCount                     int    `json:"package_files_registry_count"`
	PackageFilesVerificationTotalCount            int    `json:"package_files_verification_total_count"`
	PackageFilesVerifiedCount                     int    `json:"package_files_verified_count"`
	PackageFilesVerificationFailedCount           int    `json:"package_files_verification_failed_count"`
	PackageFilesSyncedInPercentage                string `json:"package_files_synced_in_percentage"`
	PackageFilesVerifiedInPercentage              string `json:"package_files_verified_in_percentage"`
	PagesDeploymentsCount                         int    `json:"pages_deployments_count"`
	PagesDeploymentsChecksumTotalCount            int    `json:"pages_deployments_checksum_total_count"`
	PagesDeploymentsChecksummedCount              int    `json:"pages_deployments_checksummed_count"`
	PagesDeploymentsChecksumFailedCount           int    `json:"pages_deployments_checksum_failed_count"`
	PagesDeploymentsSyncedCount                   int    `json:"pages_deployments_synced_count"`
	PagesDeploymentsFailedCount                   int    `json:"pages_deployments_failed_count"`
	PagesDeploymentsRegistryCount                 int    `json:"pages_deployments_registry_count"`
	PagesDeploymentsVerificationTotalCount        int    `json:"pages_deployments_verification_total_count"`
	PagesDeploymentsVerifiedCount                 int    `json:"pages_deployments_verified_count"`
	PagesDeploymentsVerificationFailedCount       int    `json:"pages_deployments_verification_failed_count"`
	PagesDeploymentsSyncedInPercentage            string `json:"pages_deployments_synced_in_percentage"`
	PagesDeploymentsVerifiedInPercentage          string `json:"pages_deployments_verified_in_percentage"`
	TerraformStateVersionsCount                   int    `json:"terraform_state_versions_count"`
	TerraformStateVersionsChecksumTotalCount      int    `json:"terraform_state_versions_checksum_total_count"`
	TerraformStateVersionsChecksummedCount        int    `json:"terraform_state_versions_checksummed_count"`
	TerraformStateVersionsChecksumFailedCount     int    `json:"terraform_state_versions_checksum_failed_count"`
	TerraformStateVersionsSyncedCount             int    `json:"terraform_state_versions_synced_count"`
	TerraformStateVersionsFailedCount             int    `json:"terraform_state_versions_failed_count"`
	TerraformStateVersionsRegistryCount           int    `json:"terraform_state_versions_registry_count"`
	TerraformStateVersionsVerificationTotalCount  int    `json:"terraform_state_versions_verification_total_count"`
	TerraformStateVersionsVerifiedCount           int    `json:"terraform_state_versions_verified_count"`
	TerraformStateVersionsVerificationFailedCount int    `json:"terraform_state_versions_verification_failed_count"`
	TerraformStateVersionsSyncedInPercentage      string `json:"terraform_state_versions_synced_in_percentage"`
	TerraformStateVersionsVerifiedInPercentage    string `json:"terraform_state_versions_verified_in_percentage"`
	SnippetRepositoriesCount                      int    `json:"snippet_repositories_count"`
	SnippetRepositoriesChecksumTotalCount         int    `json:"snippet_repositories_checksum_total_count"`
	SnippetRepositoriesChecksummedCount           int    `json:"snippet_repositories_checksummed_count"`
	SnippetRepositoriesChecksumFailedCount        int    `json:"snippet_repositories_checksum_failed_count"`
	SnippetRepositoriesSyncedCount                int    `json:"snippet_repositories_synced_count"`
	SnippetRepositoriesFailedCount                int    `json:"snippet_repositories_failed_count"`
	SnippetRepositoriesRegistryCount              int    `json:"snippet_repositories_registry_count"`
	SnippetRepositoriesVerificationTotalCount     int    `json:"snippet_repositories_verification_total_count"`
	SnippetRepositoriesVerifiedCount              int    `json:"snippet_repositories_verified_count"`
	SnippetRepositoriesVerificationFailedCount    int    `json:"snippet_repositories_verification_failed_count"`
	SnippetRepositoriesSyncedInPercentage         string `json:"snippet_repositories_synced_in_percentage"`
	SnippetRepositoriesVerifiedInPercentage       string `json:"snippet_repositories_verified_in_percentage"`
	GroupWikiRepositoriesCount                    int    `json:"group_wiki_repositories_count"`
	GroupWikiRepositoriesChecksumTotalCount       int    `json:"group_wiki_repositories_checksum_total_count"`
	GroupWikiRepositoriesChecksummedCount         int    `json:"group_wiki_repositories_checksummed_count"`
	GroupWikiRepositoriesChecksumFailedCount      int    `json:"group_wiki_repositories_checksum_failed_count"`
	GroupWikiRepositoriesSyncedCount              int    `json:"group_wiki_repositories_synced_count"`
	GroupWikiRepositoriesFailedCount              int    `json:"group_wiki_repositories_failed_count"`
	GroupWikiRepositoriesRegistryCount            int    `json:"group_wiki_repositories_registry_count"`
	GroupWikiRepositoriesVerificationTotalCount   int    `json:"group_wiki_repositories_verification_total_count"`
	GroupWikiRepositoriesVerifiedCount            int    `json:"group_wiki_repositories_verified_count"`
	GroupWikiRepositoriesVerificationFailedCount  int    `json:"group_wiki_repositories_verification_failed_count"`
	GroupWikiRepositoriesSyncedInPercentage       string `json:"group_wiki_repositories_synced_in_percentage"`
	GroupWikiRepositoriesVerifiedInPercentage     string `json:"group_wiki_repositories_verified_in_percentage"`
	PipelineArtifactsCount                        int    `json:"pipeline_artifacts_count"`
	PipelineArtifactsChecksumTotalCount           int    `json:"pipeline_artifacts_checksum_total_count"`
	PipelineArtifactsChecksummedCount             int    `json:"pipeline_artifacts_checksummed_count"`
	PipelineArtifactsChecksumFailedCount          int    `json:"pipeline_artifacts_checksum_failed_count"`
	PipelineArtifactsSyncedCount                  int    `json:"pipeline_artifacts_synced_count"`
	PipelineArtifactsFailedCount                  int    `json:"pipeline_artifacts_failed_count"`
	PipelineArtifactsRegistryCount                int    `json:"pipeline_artifacts_registry_count"`
	PipelineArtifactsVerificationTotalCount       int    `json:"pipeline_artifacts_verification_total_count"`
	PipelineArtifactsVerifiedCount                int    `json:"pipeline_artifacts_verified_count"`
	PipelineArtifactsVerificationFailedCount      int    `json:"pipeline_artifacts_verification_failed_count"`
	PipelineArtifactsSyncedInPercentage           string `json:"pipeline_artifacts_synced_in_percentage"`
	PipelineArtifactsVerifiedInPercentage         string `json:"pipeline_artifacts_verified_in_percentage"`
	UploadsCount                                  int    `json:"uploads_count"`
	UploadsSyncedCount                            int    `json:"uploads_synced_count"`
	UploadsFailedCount                            int    `json:"uploads_failed_count"`
	UploadsRegistryCount                          int    `json:"uploads_registry_count"`
	UploadsSyncedInPercentage                     string `json:"uploads_synced_in_percentage"`
}

// RetrieveStatusOfAllGeoNodes get the list of status of all Geo Nodes.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-status-about-all-geo-nodes
func (s *GeoNodesService) RetrieveStatusOfAllGeoNodes(options ...RequestOptionFunc) ([]*GeoNodeStatus, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "geo_nodes/status", nil, options)
	if err != nil {
		return nil, nil, err
	}

	var gnss []*GeoNodeStatus
	resp, err := s.client.Do(req, &gnss)
	if err != nil {
		return nil, resp, err
	}

	return gnss, resp, nil
}

// RetrieveStatusOfGeoNode get the of status of a specific Geo Nodes.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/geo_nodes.html#retrieve-status-about-a-specific-geo-node
func (s *GeoNodesService) RetrieveStatusOfGeoNode(id int, options ...RequestOptionFunc) (*GeoNodeStatus, *Response, error) {
	u := fmt.Sprintf("geo_nodes/%d/status", id)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gns := new(GeoNodeStatus)
	resp, err := s.client.Do(req, gns)
	if err != nil {
		return nil, resp, err
	}

	return gns, resp, nil
}
