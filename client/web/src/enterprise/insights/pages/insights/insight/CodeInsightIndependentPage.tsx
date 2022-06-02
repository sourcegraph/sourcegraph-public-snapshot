import { FunctionComponent, useContext, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, PageHeader, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../components/code-insights-page/CodeInsightsPage'
import { CodeInsightsBackendContext } from '../../../core'

import { CodeInsightIndependentPageActions } from './components/actions/CodeInsightIndependentPageActions'
import { StandaloneInsightDashboardPills } from './components/dashboard-pills/StandaloneInsightDashboardPills'
import { SmartStandaloneInsight } from './components/SmartStandaloneInsight'
import { Standalone404Insight } from './components/standalone-404-insight/Standalone404Insight'

import styles from './CodeInsightIndependentPage.module.scss'

interface CodeInsightIndependentPage extends TelemetryProps {
    insightId: string
}

export const CodeInsightIndependentPage: FunctionComponent<CodeInsightIndependentPage> = props => {
    const { insightId, telemetryService } = props
    const { getInsightById } = useContext(CodeInsightsBackendContext)

    const insight = useObservable(useMemo(() => getInsightById(insightId), [getInsightById, insightId]))

    if (insight === undefined) {
        return <LoadingSpinner inline={false} />
    }

    if (!insight) {
        return <Standalone404Insight />
    }

    return (
        <CodeInsightsPage className={styles.root}>
            <PageTitle title={`Configure ${insight.title} - Code Insights`} />
            <PageHeader
                path={[{ to: '/insights/dashboards/all', icon: CodeInsightsIcon }, { text: insight.title }]}
                actions={<CodeInsightIndependentPageActions insight={insight} />}
            />

            <StandaloneInsightDashboardPills
                dashboards={insight.dashboards}
                insightId={insight.id}
                className={styles.dashboards}
            />

            <div className={styles.content}>
                <SmartStandaloneInsight insight={insight} telemetryService={telemetryService} />
            </div>
        </CodeInsightsPage>
    )
}
