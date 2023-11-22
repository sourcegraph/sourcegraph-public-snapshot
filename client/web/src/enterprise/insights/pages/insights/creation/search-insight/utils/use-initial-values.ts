import { useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { useLocalStorage } from '@sourcegraph/wildcard'

import type { CreateInsightFormFields } from '../types'

import { decodeSearchInsightUrl } from './search-insight-url-parsers/search-insight-url-parsers'
import { useURLQueryInsight } from './use-url-query-insight/use-url-query-insight'

export interface UseInitialValuesResult {
    initialValues: Partial<CreateInsightFormFields>
    setLocalStorageFormValues: (values: CreateInsightFormFields | undefined) => void
}

export function useSearchInsightInitialValues(): UseInitialValuesResult {
    const { search } = useLocation()

    // Search insight creation UI form can take values from URL query param in order
    // to support 1-click creation insight flow for the search result page.
    const initialValuesFromURLParam = useURLQueryInsight(search)

    const urlParsedInsightValues = useMemo(() => decodeSearchInsightUrl(search), [search])

    // Creation UI saves all form values in local storage to be able to restore these
    // values if page was fully refreshed or user came back from other page.
    // We do not use temporal user settings since form values are not so important to
    // waste users time for waiting response of yet another network request to just
    // render creation UI form.
    // eslint-disable-next-line no-restricted-syntax
    const [localStorageFormValues, setLocalStorageFormValues] = useLocalStorage<CreateInsightFormFields | undefined>(
        'insights.search-insight-creation-ui-v2',
        undefined
    )

    // [1] "query" query parameter has a higher priority
    if (initialValuesFromURLParam) {
        return {
            initialValues: initialValuesFromURLParam,
            setLocalStorageFormValues,
        }
    }

    // [2] If "query" parameter isn't in URL query params then try to get encoded insight values
    if (urlParsedInsightValues) {
        return {
            initialValues: urlParsedInsightValues,
            setLocalStorageFormValues,
        }
    }

    // [3] Fallback on localstorage saved insight values
    return {
        initialValues: localStorageFormValues ?? {},
        setLocalStorageFormValues,
    }
}
