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
