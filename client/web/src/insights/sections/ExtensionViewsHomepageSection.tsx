import React, { useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { StaticView, ViewGrid } from '../../views'
import { isCodeInsightsEnabled } from '../utils/is-code-insights-enabled'

import { ExtensionViewsSectionCommonProps } from './types'

export interface ExtensionViewsHomepageSectionProps extends ExtensionViewsSectionCommonProps {
    where: 'homepage'
}

/**
 * Renders extension views section for the home (search) page. Note that this component is used only for
 * OSS version. For Sourcegraph enterprise see `./enterprise/insights/sections` components.
 */
export const ExtensionViewsHomepageSection: React.FunctionComponent<ExtensionViewsHomepageSectionProps> = props => {
    const { settingsCascade, extensionsController, className = '' } = props
    const showCodeInsights = isCodeInsightsEnabled(settingsCascade, { homepage: true })

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
        ) ?? []

    return (
        <ViewGrid
            viewIds={extensionViews.map(view => view.id)}
            telemetryService={props.telemetryService}
            className={className}
        >
            {/* Render extension views for the search page */}
            {extensionViews.map(view => (
                <StaticView key={view.id} view={view} telemetryService={props.telemetryService} />
            ))}
        </ViewGrid>
    )
}
