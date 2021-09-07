import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ViewCard } from './card/view-card/ViewCard'
import { ViewContent } from './content/view-content/ViewContent'
import { ViewErrorContent } from './content/view-error-content/ViewErrorContent'
import { ViewLoadingContent } from './content/view-loading-content/ViewLoadingContent'

interface StaticView extends TelemetryProps, React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    view: ViewProviderResult
}

/**
 * Component that renders insight-like extension card. Used by extension views in extension
 * consumers that have insight section (the search and the directory page).
 */
export const StaticView: React.FunctionComponent<StaticView> = props => {
    const { view, telemetryService, ...otherProps } = props

    return (
        <ViewCard
            data-testid={`insight-card.${view.id}`}
            insight={view}
            className="insight-content-card"
            {...otherProps}
        >
            {view.view === undefined ? (
                <ViewLoadingContent text="Loading code insight" subTitle={view.id} icon={PuzzleIcon} />
            ) : isErrorLike(view.view) ? (
                <ViewErrorContent error={view.view} title={view.id} icon={PuzzleIcon} />
            ) : (
                <ViewContent
                    telemetryService={telemetryService}
                    viewContent={view.view.content}
                    viewID={view.id}
                    containerClassName="insight-content-card"
                />
            )}
        </ViewCard>
    )
}
