import { useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { useLocalStorage } from '@sourcegraph/wildcard'

import type { CaptureGroupFormFields } from '../types'
import { decodeCaptureInsightURL } from '../utils/capture-insigh-url-parsers/capture-insight-url-parsers'

type UseCaptureInsightInitialValuesResult = [
    Partial<CaptureGroupFormFields>,
    (values: CaptureGroupFormFields | undefined) => void
]

export function useCaptureInsightInitialValues(): UseCaptureInsightInitialValuesResult {
    const { search } = useLocation()

    const urlValues = useMemo(() => decodeCaptureInsightURL(search), [search])

    // We do not use temporal user settings since form values are not so important to
    // waste users time for waiting response of yet another network request to just
    // render creation UI form.
    // eslint-disable-next-line no-restricted-syntax
    const [localStorageFormValues, setLocalStorageValues] = useLocalStorage<CaptureGroupFormFields | undefined>(
        'insights.capture-group-creation-ui-v2',
        undefined
    )

    return [urlValues ?? localStorageFormValues ?? {}, setLocalStorageValues]
}
