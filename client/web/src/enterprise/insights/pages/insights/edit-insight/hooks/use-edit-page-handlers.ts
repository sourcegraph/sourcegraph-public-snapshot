import { useContext, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { FORM_ERROR, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { Insight, isVirtualDashboard } from '../../../../core/types'
import { useQueryParameters } from '../../../../hooks/use-query-parameters'

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

    const history = useHistory()
    const { updateInsight, getDashboardById } = useContext(InsightsApiContext)
    const { dashboardId } = useQueryParameters(['dashboardId'])

    const dashboard = useObservable(useMemo(() => getDashboardById(dashboardId), [getDashboardById, dashboardId]))

    const handleSubmit = async (newInsight: Insight): Promise<SubmissionErrors> => {
        if (!originalInsight) {
            return
        }

        try {
            await updateInsight({
                oldInsight: originalInsight,
                newInsight,
            }).toPromise()

            if (!dashboard || isVirtualDashboard(dashboard)) {
                // Navigate user to the dashboard page with new created dashboard
                history.push(`/insights/dashboards/${newInsight.visibility}`)

                return
            }

            // If insight's visible area has been changed explicit redirect to new
            // scope dashboard page
            if (dashboard.owner.id !== newInsight.visibility) {
                history.push(`/insights/dashboards/${newInsight.visibility}`)
            } else {
                history.push(`/insights/dashboards/${dashboard.id}`)
            }
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => {
        history.push(`/insights/dashboards/${dashboard?.id ?? 'all'}`)
    }

    return { handleSubmit, handleCancel }
}
