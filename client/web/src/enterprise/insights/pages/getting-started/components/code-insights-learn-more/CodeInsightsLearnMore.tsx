import React, { useState } from 'react'

import { Button, Link, Popover, PopoverContent, PopoverTrigger } from '@sourcegraph/wildcard'

import { FeedbackPromptContent } from '../../../../../../nav/Feedback'

import styles from './CodeInsightsLearnMore.module.scss'

export const CodeInsightsLearnMore: React.FunctionComponent<React.HTMLAttributes<HTMLElement>> = props => {
    const [isVisible, setVisibility] = useState(false)

    return (
        <footer {...props}>
            <h2>Learn more about Code Insights</h2>

            <div className={styles.cards}>
                <article>
                    <h3>Quickstart</h3>
                    <p className="text-muted mb-2">
                        Get started and create your first code insight in 5 minutes or less.
                    </p>
                    <Link to="/help/code_insights" rel="noopener noreferrer" target="_blank">
                        Code Insights Docs
                    </Link>
                </article>

                <article>
                    <h3>Detect and track patterns</h3>
                    <p className="text-muted mb-2">
                        Track versions of languages, packages, terraform, docker images, or anything else that can be
                        captured with a regular expression capture group.
                    </p>
                    <Link
                        to="/help/code_insights/explanations/automatically_generated_data_series"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Automatically generated data series
                    </Link>
                </article>

                <article>
                    <h3>Questions and feedback</h3>
                    <p className="text-muted mb-2">
                        Have a question or idea about code monitoring? We want to hear your feedback!
                    </p>

                    <Popover isOpen={isVisible} onOpenChange={event => setVisibility(event.isOpen)}>
                        <PopoverTrigger as={Button} variant="link" className={styles.feedbackTrigger}>
                            Share your thoughts
                        </PopoverTrigger>

                        <PopoverContent className={styles.feedbackPrompt}>
                            <FeedbackPromptContent
                                closePrompt={() => setVisibility(false)}
                                textPrefix="Code Insights: "
                                routeMatch="/insights/about"
                            />
                        </PopoverContent>
                    </Popover>
                </article>
            </div>
        </footer>
    )
}
