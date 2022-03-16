import { useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { useLocalStorage } from '@sourcegraph/wildcard'

import { CaptureGroupFormFields } from '../types'
import { decodeCaptureInsightURL } from '../utils/capture-insigh-url-parsers/capture-insight-url-parsers'

type UseCaptureInsightInitialValuesResult = [
    Partial<CaptureGroupFormFields>,
    (values: CaptureGroupFormFields | undefined) => void
]

export function useCaptureInsightInitialValues(): UseCaptureInsightInitialValuesResult {
    const { search } = useLocation()

    const urlValues = useMemo(() => decodeCaptureInsightURL(search), [search])
    const [localStorageFormValues, setLocalStorageValues] = useLocalStorage<CaptureGroupFormFields | undefined>(
        'insights.capture-group-creation-ui',
        undefined
    )

    return [urlValues ?? localStorageFormValues ?? {}, setLocalStorageValues]
}
