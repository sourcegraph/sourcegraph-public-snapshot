import { FunctionComponent } from 'react'

import { PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../components/code-insights-page/CodeInsightsPage'

interface CodeInsightIndependentPage {
    insightId: string
}

export const CodeInsightIndependentPage: FunctionComponent<CodeInsightIndependentPage> = props => {
    const { insightId } = props

    return (
        <CodeInsightsPage>
            <PageTitle title={`Configure ${insightId} - Code Insights`} />

            <PageHeader path={[{ icon: CodeInsightsIcon }, { text: insightId }]} />
        </CodeInsightsPage>
    )
}
