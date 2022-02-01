import React, { useState } from 'react'

import { Button, ProductStatusBadge, Popover, PopoverTrigger, PopoverContent } from '@sourcegraph/wildcard'

import { FeedbackPromptContent } from '../../../../nav/Feedback'

import styles from './BetaFeedbackPanel.module.scss'

export const BetaFeedbackPanel: React.FunctionComponent = () => {
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center">
            <a href="https://docs.sourcegraph.com/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <ProductStatusBadge status="beta" className="text-uppercase" />
            </a>

            <Popover isOpen={isVisible} onOpenChange={event => setVisibility(event.isOpen)}>
                <PopoverTrigger as={Button} variant="link" size="sm">
                    Share feedback
                </PopoverTrigger>

                <PopoverContent className={styles.feedbackPrompt}>
                    <FeedbackPromptContent
                        closePrompt={() => setVisibility(false)}
                        textPrefix="Code Insights: "
                        routeMatch="/insights/dashboards"
                    />
                </PopoverContent>
            </Popover>
        </div>
    )
}
