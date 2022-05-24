import { FunctionComponent, useContext, useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, PageHeader, useObservable } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../../components/HeroPage'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../components/code-insights-page/CodeInsightsPage'
import { CodeInsightsBackendContext } from '../../../core'

import { CodeInsightIndependentPageActions } from './components/actions/CodeInsightIndependentPageActions'
import { SmartStandaloneInsight } from './components/SmartStandaloneInsight'

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
        return <HeroPage icon={MapSearchIcon} title="Oops, we couldn't find that insight" />
    }

    return (
        <CodeInsightsPage className={styles.root}>
            <PageTitle title={`Configure ${insight.title} - Code Insights`} />
            <PageHeader
                path={[{ to: '/insights/dashboards/all', icon: CodeInsightsIcon }, { text: insight.title }]}
                actions={<CodeInsightIndependentPageActions insight={insight} />}
            />

            <div className={styles.content}>
                <SmartStandaloneInsight insight={insight} telemetryService={telemetryService} />
            </div>
        </CodeInsightsPage>
    )
}
