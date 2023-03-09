import { gql } from '@sourcegraph/http-client'

const vulnerabilitiesFields = gql`
    fragment VulnerabilitiesFields on VulnerabilityMatch {
        __typename
        vulnerability {
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
    query VulnerabilityMatches($severity: String, $language: String) {
        vulnerabilityMatches(severity: $severity, language: $language) {
            nodes {
                ...VulnerabilitiesFields
            }
        }
    }
    ${vulnerabilitiesFields}
`
