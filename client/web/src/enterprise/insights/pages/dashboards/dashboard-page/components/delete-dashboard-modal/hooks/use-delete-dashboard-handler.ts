import { useContext, useState } from 'react'

import { ErrorLike, asError } from '@sourcegraph/common'

import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { CustomInsightDashboard } from '../../../../../../core/types'

export interface UseDeleteDashboardHandlerProps {
    dashboard: CustomInsightDashboard
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
            // if settingsKey or owner are missing then we are using the
            // new graphql api. these are not needed
            await deleteDashboard({
                dashboardSettingKey: dashboard.settingsKey || '',
                dashboardOwnerId: dashboard.owner?.id || '',
                id: dashboard.id,
            }).toPromise()

            setLoadingOrError(false)
            onSuccess()
        } catch (error: unknown) {
            setLoadingOrError(asError(error))
        }
    }

    return { handler, loadingOrError }
}
