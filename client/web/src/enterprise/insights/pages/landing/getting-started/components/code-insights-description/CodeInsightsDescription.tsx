import React from 'react'

import { H2, Text, H3, Link } from '@sourcegraph/wildcard'

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
            Code Insights gives you precise answers about the trends and composition of your codebase. It transforms
            your code into a queryable database and lets you create customizable, visual dashboards in seconds.
        </Text>

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

        <H3>Resources</H3>
        <ul>
            <li>
                <Link to="/help/code_insights" rel="noopener">
                    Code Insights documentation
                </Link>
            </li>
            <li>
                <Link to="https://about.sourcegraph.com/code-insights" rel="noopener">
                    Code Insights product page
                </Link>
            </li>
        </ul>
    </section>
)
