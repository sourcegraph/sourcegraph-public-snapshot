This is a hack since we badly vendored graphql-go with govendor. We vendored the
main package at a newer commit than the subpackages. Somehow this still worked,
and now updating to either of those commits breaks the site. This hack is
temporary until all clients are updated.
