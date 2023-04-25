"""Defines flaky tests to disable when running tests from the release we're testing the new database schema against.
"""

FLAKES = {
    "5.0.0": [
        {
            "path": "enterprise/cmd/frontend/internal/batches/resolvers",
            "prefix": "TestRepositoryPermissions",
            "reason": "Test was having incomplete data, fails now that constraints are in place"
        },
        {
            "path": "dev/sg/linters",
            "prefix": "TestLibLogLinter",
            "reason": "Test was having incomplete data, fails now that constraints are in place"
        }
    ]
}
