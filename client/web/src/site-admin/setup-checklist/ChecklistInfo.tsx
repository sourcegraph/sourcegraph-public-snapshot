import { useSearchParameters, Alert } from '@sourcegraph/wildcard'

import { withFeatureFlag } from '../../featureFlags/withFeatureFlag'

import { useSetupChecklist } from './hooks/useSetupChecklist'

export const ChecklistInfo = withFeatureFlag('setup-checklist', function ChecklistInfo() {
    const params = useSearchParameters()
    const { data, loading } = useSetupChecklist()
    if (loading) {
        return null
    }
    const paramsID = decodeURIComponent(params.get('setup-checklist') ?? '')
    const info = data.find(item => item.id === paramsID)?.info
    if (!info) {
        return null
    }

    return <Alert variant="info">{info}</Alert>
})
