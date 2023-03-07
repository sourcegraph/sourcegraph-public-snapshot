import { gql } from '@sourcegraph/http-client'

const packageRepoMatchFragment = gql`
    fragment PackageRepoMatchFields on PackageRepoReference {
        id
        name
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

export const packageRepoFilterQuery = gql`
    query PackageRepoReferencesMatchingFilter(
        $kind: PackageRepoReferenceKind!
        $filter: PackageVersionOrNameFilterInput!
        $first: Int
    ) {
        packageRepoReferencesMatchingFilter(kind: $kind, filter: $filter, first: $first) {
            nodes {
                ...PackageRepoMatchFields
            }
            totalCount
        }
    }

    ${packageRepoMatchFragment}
`

export const addPackageRepoFilterMutation = gql`
    mutation AddPackageRepoFilter(
        $kind: PackageRepoReferenceKind!
        $filter: PackageVersionOrNameFilterInput!
        $behaviour: PackageMatchBehaviour!
    ) {
        addPackageRepoFilter(kind: $kind, filter: $filter, behaviour: $behaviour) {
            alwaysNil
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
