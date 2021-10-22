import React, { useContext, useMemo } from 'react'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ExtensionViewsSectionCommonProps } from '../../../insights/sections/types'
import { isCodeInsightsEnabled } from '../../../insights/utils/is-code-insights-enabled'
import { StaticView, ViewGrid } from '../../../views'
import { SmartInsight } from '../components/insights-view-grid/components/smart-insight/SmartInsight'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../core/backend/code-insights-setting-cascade-backend'
import { Insight } from '../core/types'

export interface ExtensionViewsHomepageSectionProps extends ExtensionViewsSectionCommonProps {
    where: 'homepage'
}

const EMPTY_EXTENSION_LIST: ViewProviderResult[] = []
const EMPTY_INSIGHT_LIST: Insight[] = []

/**
 * Renders extension views section for the home (search) page. Note that this component is used only for
 * Enterprise version. For Sourcegraph OSS see `./enterprise/insights/sections` components.
 */
export const ExtensionViewsHomepageSection: React.FunctionComponent<ExtensionViewsHomepageSectionProps> = props => {
    const { platformContext, telemetryService, extensionsController, settingsCascade, className = '' } = props

    const showCodeInsights = isCodeInsightsEnabled(settingsCascade, { homepage: true })

    const api = useMemo(() => {
        console.log('recreate api context')

        return new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext)
    }, [platformContext, settingsCascade])

    if (!showCodeInsights) {
        return null
    }

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <ExtensionViewsHomepageSectionContent
                settingsCascade={settingsCascade}
                platformContext={platformContext}
                telemetryService={telemetryService}
                extensionsController={extensionsController}
                className={className}
            />
        </CodeInsightsBackendContext.Provider>
    )
}

const ExtensionViewsHomepageSectionContent: React.FunctionComponent<ExtensionViewsSectionCommonProps> = props => {
    const { extensionsController, telemetryService, className } = props
    const { getInsights } = useContext(CodeInsightsBackendContext)

    // Read insights from the setting cascade
    const insights = useObservable(useMemo(() => getInsights(), [getInsights])) ?? EMPTY_INSIGHT_LIST

    // Pull extension views by Extension API.
    const extensionViews =
        useObservable(
            useMemo(
                () =>
                    from(extensionsController.extHostAPI).pipe(
                        switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getHomepageViews({})))
                    ),
                [extensionsController]
            )
        ) ?? EMPTY_EXTENSION_LIST

    const allViewIds = useMemo(() => [...extensionViews, ...insights].map(view => view.id), [extensionViews, insights])

    return (
        <ViewGrid viewIds={allViewIds} telemetryService={telemetryService} className={className}>
            {/* Render extension views for the search page */}
            {extensionViews.map(view => (
                <StaticView key={view.id} view={view} telemetryService={telemetryService} />
            ))}
            {/* Render all code insights with proper home (search) page context */}
            {insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    where="homepage"
                    context={{}}
                />
            ))}
        </ViewGrid>
    )
}
