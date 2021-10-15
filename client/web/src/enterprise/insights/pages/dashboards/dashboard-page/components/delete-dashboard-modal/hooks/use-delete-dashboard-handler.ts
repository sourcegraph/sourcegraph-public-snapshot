import { useContext, useState } from 'react'

import { ErrorLike, asError } from '@sourcegraph/shared/src/util/errors'

import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context';
import { SettingsBasedInsightDashboard } from '../../../../../../core/types'

export interface UseDeleteDashboardHandlerProps {
    dashboard: SettingsBasedInsightDashboard
    onSuccess: () => void
}

export interface useDeleteDashboardHandlerResult {
    loadingOrError: undefined | boolean | ErrorLike
    handler: () => Promise<void>
}

/**
 * Deletes dashboard from the settings subject (owner).
 */
export function useDeleteDashboardHandler(props: UseDeleteDashboardHandlerProps): useDeleteDashboardHandlerResult {
    const { dashboard, onSuccess } = props
    const { deleteDashboard } = useContext(CodeInsightsBackendContext)

    const [loadingOrError, setLoadingOrError] = useState<undefined | boolean | ErrorLike>()

    const handler = async (): Promise<void> => {
        setLoadingOrError(true)

        try {
            await deleteDashboard({
                dashboardSettingKey: dashboard.settingsKey,
                dashboardOwnerId: dashboard.owner.id
            }).toPromise()

            setLoadingOrError(false)
            onSuccess()
        } catch (error: unknown) {
            setLoadingOrError(asError(error))
        }
    }

    return { handler, loadingOrError }
}
