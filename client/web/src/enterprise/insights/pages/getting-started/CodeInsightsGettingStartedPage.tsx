import React from 'react'

import { CodeInsightsExamples } from './components/code-insights-examples/CodeInsightsExamples'
import { CodeInsightsTemplates } from './components/code-insights-temlates/CodeInsightsTemplates'

export const CodeInsightsGettingStartedPage: React.FunctionComponent = () => (
    <div className="pb-5">
        <CodeInsightsExamples className="mt-5" />
        <CodeInsightsTemplates className="mt-5" />
    </div>
)
