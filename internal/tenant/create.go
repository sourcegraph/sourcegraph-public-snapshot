package tenant

// CreateTenant creates a new tenant. It sets up all the necessary data structures
// required for a new tenant.
func CreateTenant() error {
	// INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('recent-contributors', 'Indexes contributors in each file using repository history.', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";
	// INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('recent-views', 'Indexes users that recently viewed files in Sourcegraph.', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";
	// INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('analytics', 'Indexes ownership data to present in aggregated views like Admin > Analytics > Own and Repo > Ownership', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";
	return nil
}
