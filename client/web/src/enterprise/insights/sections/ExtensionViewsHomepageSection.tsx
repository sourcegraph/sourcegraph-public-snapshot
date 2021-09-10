import React, { useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ExtensionViewsSectionCommonProps } from '../../../insights/sections/types'
import { isCodeInsightsEnabled } from '../../../insights/utils/is-code-insights-enabled'
import { StaticView, ViewGrid } from '../../../views'
import { SmartInsight } from '../components/insights-view-grid/components/smart-insight/SmartInsight'
import { useAllInsights } from '../hooks/use-insight/use-insight'

export interface ExtensionViewsHomepageSectionProps extends ExtensionViewsSectionCommonProps {
    where: 'homepage'
}

const EMPTY_EXTENSION_LIST: ViewProviderResult[] = []

/**
 * Renders extension views section for the home (search) page. Note that this component is used only for
 * Enterprise version. For Sourcegraph OSS see `./enterprise/insights/sections` components.
 */
export const ExtensionViewsHomepageSection: React.FunctionComponent<ExtensionViewsHomepageSectionProps> = props => {
    const { telemetryService, platformContext, settingsCascade, extensionsController, className = '' } = props
    const showCodeInsights = isCodeInsightsEnabled(settingsCascade, { homepage: true })

    // Read insights from the setting cascade
    const insights = useAllInsights({ settingsCascade })

    // Pull extension views by Extension API.
    const extensionViews =
        useObservable(
            useMemo(
                () =>
                    showCodeInsights
                        ? from(extensionsController.extHostAPI).pipe(
                              switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getHomepageViews({})))
                          )
                        : EMPTY,
                [showCodeInsights, extensionsController]
            )
        ) ?? EMPTY_EXTENSION_LIST

    const allViewIds = useMemo(() => [...extensionViews, ...insights].map(view => view.id), [extensionViews, insights])

    if (!showCodeInsights) {
        return null
    }

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
                    platformContext={platformContext}
                    settingsCascade={settingsCascade}
                    where="homepage"
                    context={{}}
                />
            ))}
        </ViewGrid>
    )
}
