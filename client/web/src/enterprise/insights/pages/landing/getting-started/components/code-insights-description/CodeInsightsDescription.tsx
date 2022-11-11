import React from 'react'

import { mdiPoll, mdiOpenInNew } from '@mdi/js'

import { Card, H2, H3, H4, Icon, Text, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../../../../../tracking/eventLogger'

interface Props {
    className?: string
}

/**
 * The product description for Code Insights.
 */
export const CodeInsightsDescription: React.FunctionComponent<Props> = ({ className }) => {
    const isSourcegraphDotCom =  window.context.sourcegraphDotComMode
    
    return (
        <section className={className}>
            <H2>Track what matters in your code</H2>

            <Text>
                Code Insights provides precise answers about the trends and composition of your codebase. It transforms
                code into a queryable database and lets you create customizable, data visualization in seconds.
            </Text>
            
            {!isSourcegraphDotCom && (
                <div>
                    <H3>Use Code Insights to...</H3>

                    <ul>
                        <li>Track migrations, adoption, and deprecations</li>
                        <li>Detect versions of languages, packages, or infrastructure</li>
                        <li>Ensure removal of security vulnerabilities</li>
                        <li>Track code smells, ownership, and configurations</li>
                        <li>
                            <Link to="/help/code_insights/references/common_use_cases" rel="noopener">
                                See more use cases
                            </Link>
                        </li>
                    </ul>
                </div>
            )}

            <H3>Resources</H3>
            <ul>
                <li>
                    <Link to="/help/code_insights" rel="noopener">
                        Documentation{' '}
                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                    </Link>
                </li>
                <li>
                    <Link to="https://about.sourcegraph.com/code-insights" rel="noopener">
                        Product page{' '}
                        <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                    </Link>
                </li>
            </ul>
            
            {isSourcegraphDotCom && (
                <Card className="shadow d-flex flex-row align-items-center p-3 mt-5">
                    <Icon role="img" size="md" violetBg={true} aria-hidden={true} svgPath={mdiPoll} />
                    <div className="pl-3">
                        <H4 className="mb-1">Get insights for your code</H4>
                        <Link to="https://signup.sourcegraph.com/" onClick={() => eventLogger.log('ClickedOnCloudCTA')}>
                            Sign up for a 30-day trial on Sourcegraph Cloud.
                        </Link>
                    </div>
                </Card>
            )}
        </section>
    )
}
