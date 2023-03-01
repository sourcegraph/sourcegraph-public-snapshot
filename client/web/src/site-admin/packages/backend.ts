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

export const packageRepoMatchesQuery = gql`
    query PackageRepoMatches($scheme: PackageRepoReferenceKind!, $filter: PackageVersionOrNameMatcher!, $first: Int) {
        packageReposMatches(packageReferenceKind: $scheme, matcher: $filter, first: $first) {
            nodes {
                ...PackageRepoMatchFields
            }
            totalCount
        }
    }

    ${packageRepoMatchFragment}
`
