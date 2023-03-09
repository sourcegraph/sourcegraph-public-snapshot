import { gql } from '@sourcegraph/http-client'

const vulnerabilitiesFields = gql`
    fragment VulnerabilitiesFields on VulnerabilityMatch {
        __typename
        vulnerability {
            id
            sourceID
            details
            summary
            affectedPackages {
                packageName
                language
                versionConstraint
            }
            published
            modified
            cvssScore
            severity
        }
    }
`

export const RESOLVE_SECURITY_VULNERABILITIES_QUERY = gql`
    query VulnerabilityMatches($first: Int, $after: String, $severity: String, $language: String) {
        vulnerabilityMatches(first: $first, after: $after, severity: $severity, language: $language) {
            nodes {
                id
                ...VulnerabilitiesFields
            }
            totalCount
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
    ${vulnerabilitiesFields}
`
