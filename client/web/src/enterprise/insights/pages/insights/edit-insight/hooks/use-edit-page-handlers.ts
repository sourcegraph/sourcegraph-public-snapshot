import { useContext, useMemo } from 'react'

import { useHistory } from 'react-router-dom'

import { asError } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../../../../tracking/eventLogger'
import { FORM_ERROR, SubmissionErrors } from '../../../../components'
import { ALL_INSIGHTS_DASHBOARD } from '../../../../constants'
import { CodeInsightsBackendContext, CreationInsightInput } from '../../../../core'
import { useQueryParameters } from '../../../../hooks'
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

    const { dashboardId, insight } = useQueryParameters(['dashboardId', 'insight'])
    const dashboard = useObservable(useMemo(() => getDashboardById({ dashboardId }), [getDashboardById, dashboardId]))
    const redirectUrl = insight
        ? `/insights/insight/${insight}`
        : dashboard
        ? `/insights/dashboards/${dashboard.id}`
        : `/insights/dashboards/${ALL_INSIGHTS_DASHBOARD.id}`

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

            history.push(redirectUrl)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => {
        history.push(redirectUrl)
    }

    return { handleSubmit, handleCancel }
}
