import { gql } from '@sourcegraph/http-client'

export const RESOLVE_SECURITY_VULNERABILITIES_QUERY = gql`
    fragment SecurityVulnerabilitiesFields on Vulnerability {
        cve
        description
        dependency
        packageManager
        publishedDate
        lastUpdate
        affectedVersion
        currentVersion
        severityScore
        severityString
    }

    query Vulnerabilities($repository: ID!) {
        vulnerabilities(repository: $repository) {
            ...SecurityVulnerabilitiesFields
        }
    }
`
