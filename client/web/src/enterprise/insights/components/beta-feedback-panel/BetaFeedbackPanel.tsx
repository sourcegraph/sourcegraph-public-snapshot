import React, { useState } from 'react'

import { ProductStatusBadge, FeedbackPrompt, PopoverTrigger, Button, Link, Position } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../../../hooks'

export const BetaFeedbackPanel: React.FunctionComponent = () => {
    const [isVisible, setVisibility] = useState(false)
    const { handleSubmitFeedback } = useHandleSubmitFeedback(['/insights/dashboards'], 'Code Insights: ')

    return (
        <div className="d-flex align-items-center">
            <Link to="/help/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <ProductStatusBadge status="beta" className="text-uppercase" />
            </Link>

            <FeedbackPrompt
                position={Position.bottomStart}
                openByDefault={isVisible}
                onSubmit={handleSubmitFeedback}
                onClose={() => setVisibility(false)}
            >
                <PopoverTrigger as={Button} variant="link" size="sm">
                    Share feedback
                </PopoverTrigger>
            </FeedbackPrompt>
        </div>
    )
}
