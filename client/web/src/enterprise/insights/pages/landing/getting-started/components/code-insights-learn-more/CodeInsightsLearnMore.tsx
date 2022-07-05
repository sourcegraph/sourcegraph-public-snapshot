import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, PopoverTrigger, FeedbackPrompt, H2, H3, Text } from '@sourcegraph/wildcard'

import { useHandleSubmitFeedback } from '../../../../../../../hooks'
import { useLogEventName } from '../../../CodeInsightsLandingPageContext'

import styles from './CodeInsightsLearnMore.module.scss'

interface CodeInsightsLearnMoreProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {}

export const CodeInsightsLearnMore: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsLearnMoreProps>
> = props => {
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
            <H2>Learn more about Code Insights</H2>

            <div className={styles.cards}>
                <article>
                    <H3>Quickstart</H3>
                    <Text className="text-muted mb-2">
                        Get started and create your first code insight in 5 minutes or less.
                    </Text>
                    <Link to="/help/code_insights" rel="noopener noreferrer" target="_blank" onClick={handleLinkClick}>
                        Code Insights Docs
                    </Link>
                </article>

                <article>
                    <H3>Detect and track patterns</H3>
                    <Text className="text-muted mb-2">
                        Track versions of languages, packages, infrastructure, docker images, or anything else that can
                        be captured with a regular expression capture group.
                    </Text>
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
                    <H3>Questions and feedback</H3>
                    <Text className="text-muted mb-2">
                        Have a question or idea about Code Insights? We want to hear your feedback!
                    </Text>

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
