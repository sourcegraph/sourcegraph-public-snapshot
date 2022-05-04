import React, { useEffect } from 'react'

import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PageTitle } from '../../../../../components/PageTitle'

import { CodeInsightsExamples } from './components/code-insights-examples/CodeInsightsExamples'
import { CodeInsightsLearnMore } from './components/code-insights-learn-more/CodeInsightsLearnMore'
import { CodeInsightsTemplates } from './components/code-insights-templates/CodeInsightsTemplates'
import { DynamicCodeInsightExample } from './components/dynamic-code-insight-example/DynamicCodeInsightExample'

import styles from './CodeInsightsGettingStartedPage.module.scss'

interface CodeInsightsGettingStartedPageProps extends TelemetryProps, SettingsCascadeProps<Settings> {}

export const CodeInsightsGettingStartedPage: React.FunctionComponent<CodeInsightsGettingStartedPageProps> = props => {
    const { telemetryService } = props

    useEffect(() => {
        telemetryService.logViewEvent('InsightsGetStartedPage')
    }, [telemetryService])

    return (
        <main className="pb-5">
            <PageTitle title="Code Insights" />
            <DynamicCodeInsightExample telemetryService={telemetryService} />
            <CodeInsightsExamples telemetryService={telemetryService} className={styles.section} />
            <CodeInsightsTemplates
                telemetryService={telemetryService}
                settingsCascade={props.settingsCascade}
                className={styles.section}
            />
            <CodeInsightsLearnMore telemetryService={telemetryService} className={styles.section} />
        </main>
    )
}
