import React from 'react'

import { mdiPoll, mdiOpenInNew } from '@mdi/js'

import { Card, H2, H3, H4, Icon, Text, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../../../../../tracking/eventLogger'

import styles from './CodeInsightsDescription.module.scss'

interface Props {
    className?: string
}

/**
 * The product description for Code Insights.
 */
export const CodeInsightsDescription: React.FunctionComponent<Props> = ({ className }) => (
    <section className={className}>
        <H2>Track what matters in your code</H2>

        <Text>
            Code Insights provides precise answers about the trends and composition of your codebase. It transforms
            code into a queryable database and lets you create customizable, data visualization in seconds.
        </Text>

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
        
        <Card className="shadow d-flex flex-row align-items-center p-3 mt-5">
            <Icon role="img" size="md" className={styles.iconBackground} aria-hidden={true} svgPath={mdiPoll} />
            <div className="pl-3">
                <H4 className="mb-1">Get insights for your code</H4>
                <Link to="https://signup.sourcegraph.com/" onClick={() => eventLogger.log('ClickedOnCloudCTA')}>
                    Sign up for a 30-day trial on Sourcegraph Cloud.
                </Link>
            </div>
        </Card>
    </section>
)
