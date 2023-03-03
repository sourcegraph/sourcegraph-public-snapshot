import { ApolloError } from '@apollo/client'
import { useQuery, gql } from '@sourcegraph/http-client'
import { VulnerabilityMatchesResult } from '../../graphql-operations'

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
    query VulnerabilityMatches {
        vulnerabilityMatches {
            nodes {
                ...VulnerabilitiesFields
            }
        }
    }
    ${vulnerabilitiesFields}
`

export interface VulnerabilityMatchProps {
    vulnerabilities: VulnerabilitiesProps
}

export interface VulnerabilitiesProps {
    sourceID: string
    details: string
    summary: string
    published: string
    modified: string
    cvssScore: string
    severity: string
    affectedPackages: VulnerabilityAffectedPackage[]
}

export interface VulnerabilityAffectedPackage {
    packageName: string
    language: string
    versionConstraints: string[]
}

interface UseSentinelResult {
    vulnerabilities: VulnerabilitiesProps[]
    loading: boolean
    error: ApolloError | undefined
}

export const useSentinelQuery = (): UseSentinelResult => {
    const { data, loading, error } = useQuery<VulnerabilityMatchesResult>(RESOLVE_SECURITY_VULNERABILITIES_QUERY, {
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

    const response = data?.vulnerabilityMatches?.nodes ?? []
    const vulnerabilities: VulnerabilitiesProps[] = response.map(
        ({
            vulnerability: {
                sourceID,
                details,
                summary,
                published,
                modified = '',
                cvssScore,
                severity,
                affectedPackages,
            },
        }): VulnerabilitiesProps => ({
            sourceID,
            details,
            summary,
            published,
            modified: modified ?? '',
            cvssScore,
            severity,
            affectedPackages: affectedPackages.map(({ packageName, language, versionConstraint }) => ({
                packageName,
                language,
                versionConstraints: versionConstraint.map(constraint => constraint),
            })),
        })
    )

    console.log('🚀 ~ file: useSentinelQuery.ts:103 ~ useSentinelQuery ~ vulnerabilities:', vulnerabilities)
    return {
        vulnerabilities,
        loading,
        error,
    }
}
