import classNames from 'classnames'
import React, { useContext, useEffect } from 'react'
import { useHistory } from 'react-router'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../../../../components/Page'
import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import { BetaFeedbackPanel } from '../../../../components/beta-feedback-panel/BetaFeedbackPanel'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../../../../core/backend/gql-api/code-insights-gql-backend'

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
    const api = useContext(CodeInsightsBackendContext)

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

    const isGqlApi = api instanceof CodeInsightsGqlBackend

    return (
        <Page className={classNames('container pb-5', styles.container)}>
            <PageHeader
                annotation={<BetaFeedbackPanel />}
                path={[{ icon: CodeInsightsIcon }, { text: 'Create new code insight' }]}
                description={
                    <>
                        Insights analyze your code based on any search query.{' '}
                        <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                            Learn more
                        </a>
                    </>
                }
                className={styles.header}
            />

            <div className={styles.sectionContent}>
                <SearchInsightCard data-testid="create-search-insights" onClick={handleCreateSearchBasedInsightClick} />

                {isGqlApi && (
                    <CaptureGroupInsightCard
                        data-testid="create-capture-group-insight"
                        onClick={handleCaptureGroupInsightClick}
                    />
                )}

                <LangStatsInsightCard
                    data-testid="create-lang-usage-insight"
                    onClick={handleCreateCodeStatsInsightClick}
                />

                <div className={styles.info}>
                    Not sure which insight type to choose? Learn more about the{' '}
                    <a
                        href="https://docs.sourcegraph.com/code_insights/references/common_use_cases"
                        target="_blank"
                        rel="noopener"
                    >
                        use cases.
                    </a>
                </div>

                <ExtensionInsightsCard data-testid="explore-extensions" onClick={handleExploreExtensionsClick} />
            </div>
        </Page>
    )
}
