import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ActivationProps, percentageDone } from '../../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../../shared/src/components/activation/ActivationChecklist'
import H from 'history'

interface Props extends ActivationProps {
    history: H.History
}

/**
 * A card on the site admin overview page that displays the activation checklist.
 */
export const SiteAdminActivationChecklistOverviewCard: React.FunctionComponent<Props> = ({ activation, history }) => {
    let setupPercentage = 0
    if (activation) {
        setupPercentage = percentageDone(activation.completed)
    }

    return !activation || !activation.completed ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            <h3 className="card-header">{setupPercentage < 100 ? 'Get started' : 'Setup status'}</h3>
            <ActivationChecklist history={history} steps={activation.steps} completed={activation.completed} />
        </>
    )
}
