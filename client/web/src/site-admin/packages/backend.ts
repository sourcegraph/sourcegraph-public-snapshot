import { gql } from '@sourcegraph/http-client'

const packageRepoMatchFragment = gql`
    fragment PackageRepoMatchFields on PackageRepoReference {
        id
        name
        blocked
        repository {
            id
            url
            name
            mirrorInfo {
                byteSize
            }
        }
        versions {
            id
            version
        }
    }
`

const packageVersionMatchFragment = gql`
    fragment PackageVersionMatchFields on PackageRepoReferenceVersion {
        id
        version
    }
`

export const packageRepoFilterQuery = gql`
    query PackageRepoReferencesMatchingFilter(
        $kind: PackageRepoReferenceKind!
        $filter: PackageVersionOrNameFilterInput!
        $first: Int
    ) {
        packageRepoReferencesMatchingFilter(kind: $kind, filter: $filter, first: $first) {
            ... on PackageRepoReferenceConnection {
                nodes {
                    ...PackageRepoMatchFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                    endCursor
                }
            }
            ... on PackageRepoReferenceVersionConnection {
                nodes {
                    ...PackageVersionMatchFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                    endCursor
                }
            }
        }
    }

    ${packageRepoMatchFragment}
    ${packageVersionMatchFragment}
`

const packageRepoFilterFragment = gql`
    fragment PackageRepoFilterFields on PackageFilter {
        id
        behaviour
        kind
        nameFilter {
            packageGlob
        }
        versionFilter {
            packageName
            versionGlob
        }
    }
`

export const packageRepoFiltersQuery = gql`
    query PackageRepoFilters {
        packageRepoFilters {
            ...PackageRepoFilterFields
        }
    }

    ${packageRepoFilterFragment}
`

export const addPackageRepoFilterMutation = gql`
    mutation AddPackageRepoFilter(
        $kind: PackageRepoReferenceKind!
        $filter: PackageVersionOrNameFilterInput!
        $behaviour: PackageMatchBehaviour!
    ) {
        addPackageRepoFilter(kind: $kind, filter: $filter, behaviour: $behaviour) {
            id
        }
    }
`

export const deletePackageRepoFilterMutation = gql`
    mutation DeletePackageRepoFilter($id: ID!) {
        deletePackageRepoFilter(id: $id) {
            alwaysNil
        }
    }
`

export const updatePackageRepoFilterMutation = gql`
    mutation UpdatePackageRepoFilter(
        $id: ID!
        $kind: PackageRepoReferenceKind!
        $filter: PackageVersionOrNameFilterInput!
        $behaviour: PackageMatchBehaviour!
    ) {
        updatePackageRepoFilter(id: $id, kind: $kind, filter: $filter, behaviour: $behaviour) {
            alwaysNil
        }
    }
`
