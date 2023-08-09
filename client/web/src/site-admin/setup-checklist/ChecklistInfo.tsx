import React from 'react'

import { useSearchParameters, Alert } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

import { useSetupChecklist } from './hooks/useSetupChecklist'

export const ChecklistInfo: React.FC = () => {
    const params = useSearchParameters()
    const [isSetupChecklistEnabled] = useFeatureFlag('setup-checklist', false)
    const { data, loading } = useSetupChecklist()
    if (loading) {
        return null
    }
    if (!isSetupChecklistEnabled) {
        return null
    }
    const paramsID = decodeURIComponent(params.get('setup-checklist') ?? '')
    const info = data.find(item => item.id === paramsID)?.info
    if (!info) {
        return null
    }

    return <Alert variant="info">{info}</Alert>
}
