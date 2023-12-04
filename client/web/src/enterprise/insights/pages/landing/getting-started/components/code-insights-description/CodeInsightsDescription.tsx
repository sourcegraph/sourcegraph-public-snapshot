import React from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { H2, H3, Icon, Text, Link } from '@sourcegraph/wildcard'

interface Props {
    className?: string
}

const productPageUrl = 'https://sourcegraph.com/code-insights'

/**
 * The product description for Code Insights.
 */
export const CodeInsightsDescription: React.FunctionComponent<Props> = ({ className }) => (
    <section className={className}>
        <H2>Track what matters in your code</H2>

        <Text>
            Code Insights provides precise answers about the trends and composition of your codebase. It transforms code
            into a queryable database and lets you create customizable visual dashboards in seconds.
        </Text>

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

        <H3>Resources</H3>
        <ul>
            <li>
                <Link to="/help/code_insights" target="_blank" rel="noopener">
                    Documentation <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                </Link>
            </li>
            <li>
                <Link to={productPageUrl} target="_blank" rel="noopener">
                    Product page <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                </Link>
            </li>
        </ul>
    </section>
)
