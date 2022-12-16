import { QueryState } from '@sourcegraph/search'

import { useExperimentalFeatures } from '../../../../../../../../stores'
import {
    insightRepositoriesAsyncValidator,
    insightRepositoriesValidator,
    useField,
    useFieldAPI,
    ValidationResult,
} from '../../../../../../components'
import { FormAPI } from '../../../../../../components/form/hooks/useForm'
import { RepoMode } from '../../types'

const SEARCH_QUERY_VALIDATOR = (value?: QueryState): ValidationResult => {
    if (value && value.query.trim() === '') {
        return 'Search repositories query is a required filed, please fill in the field.'
    }
}

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
    repositories: string
}

interface Input<Fields> {
    formApi: FormAPI<Fields>
}

interface Fields {
    repoMode: useFieldAPI<RepoMode>
    repoQuery: useFieldAPI<QueryState>
    repositories: useFieldAPI<string>
}

export function useRepoFields<FormFields extends RepositoriesFields>(props: Input<FormFields>): Fields {
    const { formApi } = props
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

    const repoQuery = useField({
        formApi,
        name: 'repoQuery',
        disabled: !isSearchQueryMode,
        validators: {
            sync: isRepoQueryRequired ? SEARCH_QUERY_VALIDATOR : undefined,
        },
    })

    const repositories = useField({
        formApi,
        name: 'repositories',
        disabled: !isURLsListMode,
        required: isRepoURLsListRequired,
        validators: {
            // Turn off any validations for the repositories' field in we are in all repos mode
            sync: isRepoURLsListRequired ? insightRepositoriesValidator : undefined,
            async: isRepoURLsListRequired ? insightRepositoriesAsyncValidator : undefined,
        },
    })

    return { repoMode, repoQuery, repositories }
}
