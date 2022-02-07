import React from 'react'

import { CodeInsightsExamples } from './components/code-insights-examples/CodeInsightsExamples'
import { CodeInsightsLearnMore } from './components/code-insights-learn-more/CodeInsightsLearnMore'
import { CodeInsightsTemplates } from './components/code-insights-templates/CodeInsightsTemplates'

export const CodeInsightsGettingStartedPage: React.FunctionComponent = () => (
    <div className="pb-5">
        <CodeInsightsExamples className="mt-5" />
        <CodeInsightsTemplates className="mt-5" />
        <CodeInsightsLearnMore className="mt-5" />
    </div>
)
