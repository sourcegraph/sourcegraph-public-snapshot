import React, { useState } from 'react'

import { ProductStatusBadge, FeedbackPrompt, PopoverTrigger, Button, Link } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../../../hooks'

export const BetaFeedbackPanel: React.FunctionComponent = () => {
    const [isVisible, setVisibility] = useState(false)
    const feedbackSubmitState = useHandleSubmitFeedback({
        textPrefix: 'Code Insights: ',
        routeMatch: '/insights/dashboards',
    })

    return (
        <div className="d-flex align-items-center">
            <Link to="https://docs.sourcegraph.com/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <ProductStatusBadge status="beta" className="text-uppercase" />
            </Link>

            <FeedbackPrompt open={isVisible} {...feedbackSubmitState} closePrompt={() => setVisibility(false)}>
                <PopoverTrigger as={Button} variant="link" size="sm">
                    Share feedback
                </PopoverTrigger>
            </FeedbackPrompt>
        </div>
    )
}
