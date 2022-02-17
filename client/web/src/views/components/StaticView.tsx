import React, { forwardRef } from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import * as View from './view'

interface StaticViewProps
    extends TelemetryProps,
        React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    content: ViewProviderResult
}

/**
 * Component that renders insight-like extension card. Used by extension views in extension
 * consumers that have insight section (the search and the directory page).
 */
export const StaticView = forwardRef<HTMLElement, StaticViewProps>((props, reference) => {
    const {
        content: { view, id: contentId },
        telemetryService,
        ...otherProps
    } = props

    const title = !isErrorLike(view) ? view?.title : undefined
    const subtitle = !isErrorLike(view) ? view?.subtitle : undefined

    return (
        <View.Root
            title={title}
            subtitle={subtitle}
            className="insight-content-card"
            data-testid={`insight-card.${contentId}`}
            innerRef={reference}
            {...otherProps}
        >
            {view === undefined ? (
                <View.LoadingContent text="Loading code insight" />
            ) : isErrorLike(view) ? (
                <View.ErrorContent error={view} title={contentId} />
            ) : (
                <View.Content content={view.content} />
            )}
        </View.Root>
    )
})
