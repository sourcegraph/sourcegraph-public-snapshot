import React from 'react'

import { Button, ProductStatusBadge, PopoverTrigger, Link, FeedbackPrompt, Position } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../../../hooks'

export const BetaFeedbackPanel: React.FunctionComponent = () => {
    const { handleSubmitFeedback } = useHandleSubmitFeedback({
        routeMatch: '/insights/dashboards',
        textPrefix: 'Code Insights: ',
    })

    return (
        <div className="d-flex align-items-center">
            <Link to="/help/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <ProductStatusBadge status="beta" className="text-uppercase" />
            </Link>

            <FeedbackPrompt position={Position.bottomStart} onSubmit={handleSubmitFeedback}>
                <PopoverTrigger as={Button} variant="link" size="sm">
                    Share feedback
                </PopoverTrigger>
            </FeedbackPrompt>
        </div>
    )
}
