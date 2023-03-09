import { ApolloError, ApolloQueryResult } from '@apollo/client'
import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { RESOLVE_SECURITY_VULNERABILITIES_QUERY } from './graphqlQueries'
import {
    VulnerabilityMatchesResult,
    VulnerabilityMatchesVariables,
    VulnerabilitiesFields,
} from '../../../graphql-operations'
import {
    UseShowMorePaginationResult,
    useShowMorePagination,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'

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
    language: string
}

// export const useSentinelQuery = (args: UseSentinelProps): UseSentinelResult => {
//     const { data, loading, error, refetch } = useQuery<VulnerabilityMatchesResult, VulnerabilityMatchesVariables>(
//         RESOLVE_SECURITY_VULNERABILITIES_QUERY,
//         {
//             variables: {
//                 severity: args.severity,
//                 language: args.language,
//             },
//             notifyOnNetworkStatusChange: false,
//             fetchPolicy: 'no-cache',
//         }
//     )

//     const response = data?.vulnerabilityMatches?.nodes ?? []
//     const vulnerabilities: VulnerabilitiesProps[] = response.map(
//         ({
//             vulnerability: {
//                 sourceID,
//                 details,
//                 summary,
//                 published,
//                 modified = '',
//                 cvssScore,
//                 severity,
//                 affectedPackages,
//             },
//         }): VulnerabilitiesProps => ({
//             sourceID,
//             details,
//             summary,
//             published,
//             modified: modified ?? '',
//             cvssScore,
//             severity,
//             affectedPackages: affectedPackages.map(({ packageName, language, versionConstraint }) => ({
//                 packageName,
//                 language,
//                 versionConstraints: versionConstraint.map(constraint => constraint),
//             })),
//         })
//     )

//     console.log('🚀 ~ file: useSentinelQuery.ts:103 ~ useSentinelQuery ~ vulnerabilities:', vulnerabilities)
//     return {
//         vulnerabilities,
//         loading,
//         error,
//         refetch,
//     }
// }

export const useSentinelQuery = (
    args: UseSentinelProps
): UseShowMorePaginationResult<VulnerabilityMatchesResult, VulnerabilitiesFields> => {
    return useShowMorePagination<VulnerabilityMatchesResult, VulnerabilityMatchesVariables, VulnerabilitiesFields>({
        query: RESOLVE_SECURITY_VULNERABILITIES_QUERY,
        variables: {
            after: null,
            first: 50,
            severity: args.severity,
            language: args.language,
        },
        options: {
            fetchPolicy: 'no-cache',
        },
        getConnection: result => dataOrThrowErrors(result).vulnerabilityMatches,
    })
}

//     RESOLVE_SECURITY_VULNERABILITIES_QUERY,
//     {
//         variables: {
//             severity: args.severity,
//             language: args.language,
//         },
//         notifyOnNetworkStatusChange: false,
//         fetchPolicy: 'no-cache',
//     }
// )

// const response = data?.vulnerabilityMatches?.nodes ?? []
// const vulnerabilities: VulnerabilitiesProps[] = response.map(
//     ({
//         vulnerability: {
//             sourceID,
//             details,
//             summary,
//             published,
//             modified = '',
//             cvssScore,
//             severity,
//             affectedPackages,
//         },
//     }): VulnerabilitiesProps => ({
//         sourceID,
//         details,
//         summary,
//         published,
//         modified: modified ?? '',
//         cvssScore,
//         severity,
//         affectedPackages: affectedPackages.map(({ packageName, language, versionConstraint }) => ({
//             packageName,
//             language,
//             versionConstraints: versionConstraint.map(constraint => constraint),
//         })),
//     })
// )

// console.log('🚀 ~ file: useSentinelQuery.ts:103 ~ useSentinelQuery ~ vulnerabilities:', vulnerabilities)
// return {
//     vulnerabilities,
//     loading,
//     error,
//     refetch,
// }
// }
