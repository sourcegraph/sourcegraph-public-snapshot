import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, PopoverTrigger, FeedbackPrompt } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../../../../../../hooks'
import { useLogEventName } from '../../../CodeInsightsLandingPageContext'

import styles from './CodeInsightsLearnMore.module.scss'

interface CodeInsightsLearnMoreProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {}

export const CodeInsightsLearnMore: React.FunctionComponent<CodeInsightsLearnMoreProps> = props => {
    const { telemetryService, ...otherProps } = props
    const textDocumentClickPingName = useLogEventName('InsightsGetStartedDocsClicks')

    const { handleSubmitFeedback } = useHandleSubmitFeedback({
        routeMatch: '/insights/about',
        textPrefix: 'Code Insights: ',
    })

    const handleLinkClick = (): void => {
        telemetryService.log(textDocumentClickPingName)
    }

    return (
        <footer {...otherProps}>
            <h2>Learn more about Code Insights</h2>

            <div className={styles.cards}>
                <article>
                    <h3>Quickstart</h3>
                    <p className="text-muted mb-2">
                        Get started and create your first code insight in 5 minutes or less.
                    </p>
                    <Link to="/help/code_insights" rel="noopener noreferrer" target="_blank" onClick={handleLinkClick}>
                        Code Insights Docs
                    </Link>
                </article>

                <article>
                    <h3>Detect and track patterns</h3>
                    <p className="text-muted mb-2">
                        Track versions of languages, packages, infrastructure, docker images, or anything else that can
                        be captured with a regular expression capture group.
                    </p>
                    <Link
                        to="/help/code_insights/explanations/automatically_generated_data_series"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={handleLinkClick}
                    >
                        Automatically generated data series
                    </Link>
                </article>

                <article>
                    <h3>Questions and feedback</h3>
                    <p className="text-muted mb-2">
                        Have a question or idea about Code Insights? We want to hear your feedback!
                    </p>

                    <FeedbackPrompt onSubmit={handleSubmitFeedback}>
                        <PopoverTrigger as={Button} variant="link" className={styles.feedbackTrigger}>
                            Share your thoughts
                        </PopoverTrigger>
                    </FeedbackPrompt>
                </article>
            </div>
        </footer>
    )
}
