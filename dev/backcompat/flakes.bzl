"""Defines flaky tests to disable when running tests from the release we're testing the new database schema against.
"""

FLAKES = {
    "5.0.0": [
        {
            "path": "enterprise/cmd/frontend/internal/batches/resolvers",
            "prefix": "TestRepositoryPermissions",
            "reason": "Test was having incomplete data, fails now that constraints are in place",
        },
        {
            "path": "dev/sg/linters",
            "prefix": "TestLibLogLinter",
            "reason": "Test was having incomplete data, fails now that constraints are in place",
        },
        {
            "path": "internal/database",
            "prefix": "TestRepos_List_LastChanged",
            "reason": "Shifting constraints on table; ranking is experimental",
        },
        {
            "path": "internal/codeintel/ranking/internal/store",
            "prefix": "Test",
            "reason": "Shifting constraints on table; ranking is experimental",
        },
    ],
    "5.1.0": [],
}
