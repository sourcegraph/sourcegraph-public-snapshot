import { useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { isErrorLike } from '@sourcegraph/common'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { CreateInsightFormFields } from '../types'

import { decodeSearchInsightUrl } from './search-insight-url-parsers/search-insight-url-parsers'
import { useURLQueryInsight } from './use-url-query-insight/use-url-query-insight'

export interface UseInitialValuesResult {
    initialValues: Partial<CreateInsightFormFields>
    loading: boolean
    setLocalStorageFormValues: (values: CreateInsightFormFields | undefined) => void
}

export function useSearchInsightInitialValues(): UseInitialValuesResult {
    const { search } = useLocation()

    // Search insight creation UI form can take values from URL query param in order
    // to support 1-click creation insight flow for the search result page.
    const { hasQueryInsight, data: urlQueryInsightValues } = useURLQueryInsight(search)

    const urlParsedInsightValues = useMemo(() => decodeSearchInsightUrl(search), [search])

    // Creation UI saves all form values in local storage to be able restore these
    // values if page was fully refreshed or user came back from other page.
    const [localStorageFormValues, setLocalStorageFormValues] = useLocalStorage<CreateInsightFormFields | undefined>(
        'insights.search-insight-creation-ui',
        undefined
    )

    // [1] "query" query parameter has a higher priority
    if (hasQueryInsight) {
        return {
            initialValues: !isErrorLike(urlQueryInsightValues) ? urlQueryInsightValues ?? {} : {},
            loading: urlQueryInsightValues === undefined,
            setLocalStorageFormValues,
        }
    }

    // [2] If "query" parameter isn't in URL query params then try to get encoded insight values
    if (urlParsedInsightValues) {
        return {
            initialValues: urlParsedInsightValues,
            loading: false,
            setLocalStorageFormValues,
        }
    }

    // [3] Fallback on localstorage saved insight values
    return {
        initialValues: localStorageFormValues ?? {},
        loading: false,
        setLocalStorageFormValues,
    }
}
