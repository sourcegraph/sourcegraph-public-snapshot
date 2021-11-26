import React, { useMemo } from 'react'
import { combineLatest } from 'rxjs'
import { map } from 'rxjs/operators'

import { isDefined } from '@sourcegraph/common'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { ViewGrid } from '../../../../views'
import { SmartInsight } from '../../../insights/components/insights-view-grid/components/smart-insight/SmartInsight'
import { CodeInsightsBackendContext } from '../../../insights/core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../insights/core/backend/setting-based-api/code-insights-setting-cascade-backend'

interface Props extends TelemetryProps, SettingsCascadeProps, PlatformContextProps {
    component: Scalars['ID']
    className?: string
}

const INSIGHT_IDS = [
    'searchInsights.insight.isLightThemeThemePropsMigration',
    'searchInsights.insight.flexboxVsGrid',
    'searchInsights.insight.todOs',
]
const ENTITY_INSIGHTS_PAGE_CONTEXT = {}

export const ComponentInsights: React.FunctionComponent<Props> = ({
    settingsCascade,
    platformContext,
    telemetryService,
    className,
}) => {
    const api = useMemo(() => new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext), [
        platformContext,
        settingsCascade,
    ])

    const insights = useObservable(
        useMemo(
            () =>
                combineLatest(INSIGHT_IDS.map(id => api.getInsightById(id))).pipe(
                    map(insights => insights.filter(isDefined))
                ),
            [api]
        )
    )

    return insights ? (
        <CodeInsightsBackendContext.Provider value={api}>
            <ViewGrid viewIds={insights.map(({ id }) => id)} telemetryService={telemetryService} className={className}>
                {insights.map(insight => (
                    <SmartInsight
                        key={insight.id}
                        insight={insight}
                        telemetryService={telemetryService}
                        where="insightsPage" // TODO(sqs): add new 'where'
                        context={ENTITY_INSIGHTS_PAGE_CONTEXT}
                    />
                ))}
            </ViewGrid>
        </CodeInsightsBackendContext.Provider>
    ) : (
        <LoadingSpinner className={className + ' d-none'} />
    )
}
