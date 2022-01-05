import React, { useRef, useState } from 'react'

import { Popover, Button, ProductStatusBadge } from '@sourcegraph/wildcard'
import { flipRightPosition } from '@sourcegraph/wildcard/src/components/Popover'

import { FeedbackPromptContent } from '../../../../nav/Feedback/FeedbackPrompt'

import styles from './BetaFeedbackPanel.module.scss'

export const BetaFeedbackPanel: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center">
            <a href="https://docs.sourcegraph.com/code_insights#code-insights-beta" target="_blank" rel="noopener">
                <ProductStatusBadge status="beta" className="text-uppercase" />
            </a>

            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>

            <Popover
                isOpen={isVisible}
                target={buttonReference}
                position={flipRightPosition}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Code Insights: "
                    routeMatch="/insights/dashboards"
                />
            </Popover>
        </div>
    )
}
