import { useContext, useMemo } from 'react'

import { useHistory } from 'react-router-dom'

import { asError } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../../../../tracking/eventLogger'
import { FORM_ERROR, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CreationInsightInput } from '../../../../core/backend/code-insights-backend-types'
import { ALL_INSIGHTS_DASHBOARD } from '../../../../core/constants'
import { useQueryParameters } from '../../../../hooks/use-query-parameters'
import { getTrackingTypeByInsightType } from '../../../../pings'

export interface useHandleSubmitOutput {
    handleSubmit: (newInsight: CreationInsightInput) => Promise<SubmissionErrors>
    handleCancel: () => void
}

/**
 * Returns submit and cancel handlers for the insight edit submit page.
 */
export function useEditPageHandlers(props: { id: string | undefined }): useHandleSubmitOutput {
    const { id } = props
    const { updateInsight, getDashboardById } = useContext(CodeInsightsBackendContext)
    const history = useHistory()

    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useObservable(useMemo(() => getDashboardById({ dashboardId }), [getDashboardById, dashboardId]))

    const handleSubmit = async (newInsight: CreationInsightInput): Promise<SubmissionErrors> => {
        if (!id) {
            return
        }

        try {
            await updateInsight({
                insightId: id,
                nextInsightData: newInsight,
            }).toPromise()

            const insightType = getTrackingTypeByInsightType(newInsight.type)

            eventLogger.log('InsightEdit', { insightType }, { insightType })

            if (!dashboard) {
                // Navigate user to the dashboard page with new created dashboard
                history.push(`/insights/dashboards/${ALL_INSIGHTS_DASHBOARD.id}`)

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
            history.push(`/insights/dashboards/${ALL_INSIGHTS_DASHBOARD.id}`)
            return
        }

        history.push(`/insights/dashboards/${dashboard.id}`)
    }

    return { handleSubmit, handleCancel }
}
