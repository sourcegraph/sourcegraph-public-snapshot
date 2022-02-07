import React from 'react'

import styles from './CodeInsightsGettingStartedPage.module.scss'
import { CodeInsightsExamples } from './components/code-insights-examples/CodeInsightsExamples'
import { CodeInsightsLearnMore } from './components/code-insights-learn-more/CodeInsightsLearnMore'
import { CodeInsightsTemplates } from './components/code-insights-templates/CodeInsightsTemplates'
import { DynamicCodeInsightExample } from './components/dynamic-code-insight-example/DynamicCodeInsightExample'

export const CodeInsightsGettingStartedPage: React.FunctionComponent = () => (
    <main className="pb-5">
        <DynamicCodeInsightExample />
        <CodeInsightsExamples className={styles.section} />
        <CodeInsightsTemplates className={styles.section} />
        <CodeInsightsLearnMore className={styles.section} />
    </main>
)
