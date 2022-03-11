import React, { useEffect } from 'react'
import { useHistory } from 'react-router'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Link } from '@sourcegraph/wildcard'

import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../../components/code-insights-page/CodeInsightsPage'

import {
    CaptureGroupInsightCard,
    ExtensionInsightsCard,
    LangStatsInsightCard,
    SearchInsightCard,
} from './cards/InsightCards'
import styles from './IntroCreationPage.module.scss'

interface IntroCreationPageProps extends TelemetryProps {}

/** Displays intro page for insights creation UI. */
export const IntroCreationPage: React.FunctionComponent<IntroCreationPageProps> = props => {
    const { telemetryService } = props

    const history = useHistory()
    const { search } = useLocation()

    const handleCreateSearchBasedInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateSearchBasedInsightClick')
        history.push(`/insights/create/search${search}`)
    }

    const handleCaptureGroupInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCaptureGroupInsightClick')
        history.push(`/insights/create/capture-group${search}`)
    }

    const handleCreateCodeStatsInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCodeStatsInsightClick')
        history.push(`/insights/create/lang-stats${search}`)
    }

    const handleExploreExtensionsClick = (): void => {
        telemetryService.log('CodeInsightsExploreInsightExtensionsClick')
        history.push('/extensions?query=category:Insights&experimental=true')
    }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCreationPage')
    }, [telemetryService])

    return (
        <CodeInsightsPage className={styles.container}>
            <PageHeader
                path={[{ icon: CodeInsightsIcon }, { text: 'Create new code insight' }]}
                description={
                    <>
                        Insights analyze your code based on any search query.{' '}
                        <Link to="/help/code_insights" target="_blank" rel="noopener">
                            Learn more
                        </Link>
                    </>
                }
                className={styles.header}
            />

            <div className={styles.sectionContent}>
                <SearchInsightCard data-testid="create-search-insights" onClick={handleCreateSearchBasedInsightClick} />

                <CaptureGroupInsightCard
                    data-testid="create-capture-group-insight"
                    onClick={handleCaptureGroupInsightClick}
                />

                <LangStatsInsightCard
                    data-testid="create-lang-usage-insight"
                    onClick={handleCreateCodeStatsInsightClick}
                />

                <div className={styles.info}>
                    Not sure which insight type to choose? Learn more about the{' '}
                    <Link to="/help/code_insights/references/common_use_cases" target="_blank" rel="noopener">
                        use cases.
                    </Link>
                </div>

                <ExtensionInsightsCard data-testid="explore-extensions" onClick={handleExploreExtensionsClick} />
            </div>
        </CodeInsightsPage>
    )
}
