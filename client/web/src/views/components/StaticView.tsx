import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import * as View from './view'

interface StaticView extends TelemetryProps, React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    content: ViewProviderResult
}

/**
 * Component that renders insight-like extension card. Used by extension views in extension
 * consumers that have insight section (the search and the directory page).
 */
export const StaticView: React.FunctionComponent<StaticView> = props => {
    const { content, telemetryService, ...otherProps } = props
    const view = content.view

    const title = !isErrorLike(view) ? view?.title : undefined
    const subtitle = !isErrorLike(view) ? view?.subtitle : undefined

    return (
        <View.Root
            title={title}
            subtitle={subtitle}
            className="insight-content-card"
            data-testid={`insight-card.${content.id}`}
            {...otherProps}
        >
            {view === undefined ? (
                <View.LoadingContent text="Loading code insight" description={content.id} icon={PuzzleIcon} />
            ) : isErrorLike(view) ? (
                <View.ErrorContent error={view} title={content.id} icon={PuzzleIcon} />
            ) : (
                <View.Content
                    telemetryService={telemetryService}
                    content={view.content}
                    viewID={content.id}
                    containerClassName="insight-content-card"
                />
            )}
        </View.Root>
    )
}
