import { dataOrThrowErrors } from '@sourcegraph/http-client'

import {
    UseShowMorePaginationResult,
    useShowMorePagination,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    VulnerabilityMatchesResult,
    VulnerabilityMatchesVariables,
    VulnerabilitiesFields,
} from '../../../graphql-operations'

import { RESOLVE_SECURITY_VULNERABILITIES_QUERY } from './graphqlQueries'

interface UseSentinelProps {
    severity: string
    language: string
    repositoryName: string
}

export const useSentinelQuery = ({
    severity,
    language,
    repositoryName,
}: UseSentinelProps): UseShowMorePaginationResult<VulnerabilityMatchesResult, VulnerabilitiesFields> => {
    return useShowMorePagination<VulnerabilityMatchesResult, VulnerabilityMatchesVariables, VulnerabilitiesFields>({
        query: RESOLVE_SECURITY_VULNERABILITIES_QUERY,
        variables: { after: null, first: 50, severity, language, repositoryName },
        options: { fetchPolicy: 'network-only' },
        getConnection: result => {
            const { vulnerabilityMatches } = dataOrThrowErrors(result)
            return vulnerabilityMatches
        },
    })
}
