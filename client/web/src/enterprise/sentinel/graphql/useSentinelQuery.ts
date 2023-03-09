import { ApolloError, ApolloQueryResult } from '@apollo/client'
import { useQuery } from '@sourcegraph/http-client'
import { RESOLVE_SECURITY_VULNERABILITIES_QUERY } from './graphqlQueries'
import { VulnerabilityMatchesResult, VulnerabilityMatchesVariables } from '../../../graphql-operations'

// const vulnerabilitiesFields = gql`
//     fragment VulnerabilitiesFields on VulnerabilityMatch {
//         __typename
//         vulnerability {
//             sourceID
//             details
//             summary
//             affectedPackages {
//                 packageName
//                 language
//                 versionConstraint
//             }
//             published
//             modified
//             cvssScore
//             severity
//         }
//     }
// `

// export const RESOLVE_SECURITY_VULNERABILITIES_QUERY = gql`
//     query VulnerabilityMatches($severity: String) {
//         vulnerabilityMatches(severity: $severity) {
//             nodes {
//                 ...VulnerabilitiesFields
//             }
//         }
//     }
//     ${vulnerabilitiesFields}
// `

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
    refetch: (
        variables?: Partial<VulnerabilityMatchesVariables> | undefined
    ) => Promise<ApolloQueryResult<VulnerabilityMatchesResult>>
}

interface UseSentinelProps {
    severity: string
}

export const useSentinelQuery = (args: UseSentinelProps): UseSentinelResult => {
    const { data, loading, error, refetch } = useQuery<VulnerabilityMatchesResult, VulnerabilityMatchesVariables>(
        RESOLVE_SECURITY_VULNERABILITIES_QUERY,
        {
            variables: {
                severity: args.severity,
            },
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

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
        refetch,
    }
}
