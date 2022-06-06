package authz

//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/authz -i SubRepoPermissionChecker -o mock_sub_repo_perms_checker.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/authz -i SubRepoPermissionsGetter -o mock_sub_repo_perms_getter.go
