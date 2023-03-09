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

interface UseSentinelProps {
    severity: string
    language: string
}

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
