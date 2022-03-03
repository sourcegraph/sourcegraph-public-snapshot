import { useContext, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { asError } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../../../../tracking/eventLogger'
import { FORM_ERROR, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { Insight } from '../../../../core/types'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../../../core/types/dashboard/virtual-dashboard'
import { useQueryParameters } from '../../../../hooks/use-query-parameters'
import { getTrackingTypeByInsightType } from '../../../../pings'

export interface UseHandleSubmitProps {
    originalInsight: Insight | null | undefined
}

export interface useHandleSubmitOutput {
    handleSubmit: (newInsight: Insight) => Promise<SubmissionErrors>
    handleCancel: () => void
}

/**
 * Returns submit and cancel handlers for the insight edit submit page.
 */
export function useEditPageHandlers(props: UseHandleSubmitProps): useHandleSubmitOutput {
    const { originalInsight } = props

    const { updateInsight, getDashboardById } = useContext(CodeInsightsBackendContext)
    const history = useHistory()

    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useObservable(useMemo(() => getDashboardById({ dashboardId }), [getDashboardById, dashboardId]))

    const handleSubmit = async (newInsight: Insight): Promise<SubmissionErrors> => {
        if (!originalInsight) {
            return
        }

        try {
            await updateInsight({
                oldInsight: originalInsight,
                newInsight,
            }).toPromise()

            const insightType = getTrackingTypeByInsightType(newInsight.viewType)

            eventLogger.log('InsightEdit', { insightType }, { insightType })

            if (!dashboard) {
                // Navigate user to the dashboard page with new created dashboard
                history.push(`/insights/dashboards/${ALL_INSIGHTS_DASHBOARD_ID}`)

                return
            }

            history.push(`/insights/dashboards/${dashboard.id}`)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => {
        if (!dashboard) {
            history.push(`/insights/dashboards/${ALL_INSIGHTS_DASHBOARD_ID}`)
            return
        }

        history.push(`/insights/dashboards/${dashboard.id}`)
    }

    return { handleSubmit, handleCancel }
}
