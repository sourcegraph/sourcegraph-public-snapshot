import { useContext } from 'react'

import { useHistory } from 'react-router-dom'

import { eventLogger } from '../../../../../../tracking/eventLogger'
import { SubmissionErrors } from '../../../../components'
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
    const { updateInsight } = useContext(CodeInsightsBackendContext)
    const history = useHistory()

    const { dashboardId, insight } = useQueryParameters(['dashboardId', 'insight'])
    const redirectUrl = getReturnToLink(insight, dashboardId)

    const handleSubmit = async (newInsight: CreationInsightInput): Promise<SubmissionErrors> => {
        if (!id) {
            return
        }

        await updateInsight({
            insightId: id,
            nextInsightData: newInsight,
        }).toPromise()

        const insightType = getTrackingTypeByInsightType(newInsight.type)
        eventLogger.log('InsightEdit', { insightType }, { insightType })
        history.push(redirectUrl)
    }

    const handleCancel = (): void => {
        history.push(redirectUrl)
    }

    return { handleSubmit, handleCancel }
}

function getReturnToLink(insightId: string | undefined, dashboardId: string | undefined): string {
    if (insightId) {
        return `/insights/insight/${insightId}`
    }

    if (dashboardId) {
        return `/insights/dashboards/${dashboardId}`
    }

    return '/insights/all'
}
