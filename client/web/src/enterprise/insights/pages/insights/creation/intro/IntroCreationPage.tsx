import React, { useEffect } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../../components'

import {
    CaptureGroupInsightCard,
    ComputeInsightCard,
    LangStatsInsightCard,
    SearchInsightCard,
} from './cards/InsightCards'

import styles from './IntroCreationPage.module.scss'

interface IntroCreationPageProps extends TelemetryProps {}

/** Displays intro page for insights creation UI. */
export const IntroCreationPage: React.FunctionComponent<React.PropsWithChildren<IntroCreationPageProps>> = props => {
    const { telemetryService } = props

    const navigate = useNavigate()
    const { search } = useLocation()
    const codeInsightsCompute = useExperimentalFeatures(features => features.codeInsightsCompute)

    const handleCreateSearchBasedInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateSearchBasedInsightClick')
        navigate(`/insights/create/search${search}`)
    }

    const handleCaptureGroupInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCaptureGroupInsightClick')
        navigate(`/insights/create/capture-group${search}`)
    }

    const handleCreateComputeInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateComputeInsightClick')
        navigate(`/insights/create/group-results${search}`)
    }

    const handleCreateCodeStatsInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCodeStatsInsightClick')
        navigate(`/insights/create/lang-stats${search}`)
    }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCreationPage')
    }, [telemetryService])

    return (
        <CodeInsightsPage className={styles.container}>
            <PageTitle title="Create insight - Code Insights" />
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
                <SearchInsightCard
                    data-testid="create-search-insights"
                    handleCreate={handleCreateSearchBasedInsightClick}
                />

                <CaptureGroupInsightCard
                    data-testid="create-capture-group-insight"
                    handleCreate={handleCaptureGroupInsightClick}
                />

                {codeInsightsCompute && (
                    <ComputeInsightCard
                        data-testid="create-compute-insights"
                        handleCreate={handleCreateComputeInsightClick}
                    />
                )}

                <LangStatsInsightCard
                    data-testid="create-lang-usage-insight"
                    handleCreate={handleCreateCodeStatsInsightClick}
                />

                <div className={styles.info}>
                    Not sure which insight type to choose? Learn more about the{' '}
                    <Link to="/help/code_insights/references/common_use_cases" target="_blank" rel="noopener">
                        use cases.
                    </Link>
                </div>
            </div>
        </CodeInsightsPage>
    )
}
