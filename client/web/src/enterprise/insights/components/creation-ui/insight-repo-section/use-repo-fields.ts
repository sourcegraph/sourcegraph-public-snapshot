import { useMemo } from 'react'

import { type ApolloClient, gql, useApolloClient } from '@apollo/client'

import type { QueryState } from '@sourcegraph/shared/src/search'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import {
    type FormAPI,
    type AsyncValidator,
    useField,
    type useFieldAPI,
    type ValidationResult,
} from '@sourcegraph/wildcard'

import type {
    ValidateInsightRepoQueryResult,
    ValidateInsightRepoQueryVariables,
} from '../../../../../graphql-operations'
import type { RepoMode } from '../../../pages/insights/creation/search-insight/types'
import { insightRepositoriesValidator } from '../validators/validators'

interface RepositoriesFields {
    /**
     * [Experimental] Repositories UI can work in different modes when we have
     * two repo UI fields version of the creation UI. This field controls the
     * current mode
     */
    repoMode: RepoMode

    /**
     * Search-powered query, this is used to gather different repositories though
     * search API instead of having strict list of repo URLs.
     */
    repoQuery: QueryState

    /** Repositories which to be used to get the info for code insights */
    repositories: string[]
}

interface Input<Fields> {
    formApi: FormAPI<Fields>
}

interface Fields {
    repoMode: useFieldAPI<RepoMode>
    repoQuery: useFieldAPI<QueryState>
    repositories: useFieldAPI<string[]>
}

export function useRepoFields<FormFields extends RepositoriesFields>(props: Input<FormFields>): Fields {
    const { formApi } = props

    const apolloClient = useApolloClient()
    const repoFieldVariation = useExperimentalFeatures(features => features.codeInsightsRepoUI)

    const isSingleSearchQueryRepo = repoFieldVariation === 'single-search-query'
    const isSearchQueryORUrlsList = repoFieldVariation === 'search-query-or-strict-list'

    const repoMode = useField({
        formApi,
        name: 'repoMode',
    })

    const isSearchQueryMode = repoMode.meta.value === 'search-query'
    const isURLsListMode = repoMode.meta.value === 'urls-list'

    // Search query field is required only if it's only one option for the filling in
    // repositories info (in case of "single-search-query" UI variation) or when
    // we are in the "search-query" repo mode (in case of "search-query-or-strict-list" UI variation)
    const isRepoQueryRequired = isSingleSearchQueryRepo || isSearchQueryMode

    // Repo urls list field is required only if we are in the "search-query-or-strict-list" UI variation,
    // and we picked urls-list repo mode in the UI. In all other cases this field nighter rendered nor
    // required
    const isRepoURLsListRequired = isSearchQueryORUrlsList && isURLsListMode

    const validateRepoQuerySyntax = useMemo(() => createValidateRepoQuerySyntax(apolloClient), [apolloClient])

    const repoQuery = useField({
        formApi,
        name: 'repoQuery',
        disabled: !isSearchQueryMode,
        validators: {
            sync: isRepoQueryRequired ? validateRepoQuery : undefined,
            async: isRepoQueryRequired ? validateRepoQuerySyntax : undefined,
        },
    })

    const repositories = useField({
        formApi,
        name: 'repositories',
        disabled: !isURLsListMode,
        validators: {
            // Turn off any validations for the repositories' field in we are in all repos mode
            sync: isRepoURLsListRequired ? insightRepositoriesValidator : undefined,
        },
    })

    return { repoMode, repoQuery, repositories }
}

function validateRepoQuery(value?: QueryState): ValidationResult {
    if (value && value.query.trim() === '') {
        return 'Search repositories query is a required field, please fill in the field.'
    }
}

const VALIDATE_REPO_QUERY_GQL = gql`
    query ValidateInsightRepoQuery($query: String!) {
        validateScopedInsightQuery(query: $query) {
            isValid
            invalidReason
        }
    }
`

function createValidateRepoQuerySyntax(apolloClient: ApolloClient<unknown>): AsyncValidator<QueryState> {
    return async (value?: QueryState): Promise<ValidationResult<unknown>> => {
        if (!value) {
            return
        }

        const { data } = await apolloClient.query<ValidateInsightRepoQueryResult, ValidateInsightRepoQueryVariables>({
            query: VALIDATE_REPO_QUERY_GQL,
            variables: { query: value.query },
        })

        if (data.validateScopedInsightQuery.invalidReason) {
            return data.validateScopedInsightQuery.invalidReason
        }
    }
}
