import React from 'react'

import { CodeInsightsExamples } from './components/code-insights-examples/CodeInsightsExamples'
import { DynamicCodeInsightExample } from './components/dynamic-code-insight-example/DynamicCodeInsightExample'

export const CodeInsightsGettingStartedPage: React.FunctionComponent = () => (
    <div>
        <DynamicCodeInsightExample />
        <CodeInsightsExamples />
    </div>
)
