import React, { useEffect, useRef } from 'react'
import { Markdown } from '../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import * as H from 'history'
import { QueryInputInViewContent } from './QueryInputInViewContent'
import { View, MarkupContent } from 'sourcegraph'
import { CaseSensitivityProps, PatternTypeProps, CopyQueryButtonProps, SearchContextProps } from '../search'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { hasProperty } from '../../../shared/src/util/types'
import { isObject } from 'lodash'
import { VersionContextProps } from '../../../shared/src/search/util'
import { ChartViewContent } from './ChartViewContent'
import { TelemetryProps } from '../../../shared/src/telemetry/telemetryService'

const isMarkupContent = (input: unknown): input is MarkupContent =>
    isObject(input) && hasProperty('value')(input) && typeof input.value === 'string'

export interface ViewContentProps
    extends SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        TelemetryProps {
    viewContent: View['content']
    viewID: string
    location: H.Location
    history: H.History
    globbing: boolean

    /** To get container to track hovers for pings */
    containerClassName?: string
}

/**
 * Renders the content of an extension-contributed view.
 */
export const ViewContent: React.FunctionComponent<ViewContentProps> = ({
    viewContent,
    viewID,
    containerClassName,
    ...props
}) => {
    // Track user intent to interact with extension-contributed views
    const viewContentReference = useRef<HTMLDivElement>(null)
    useEffect(() => {
        let viewContentElement = viewContentReference.current

        let timeoutID: number | undefined

        function onMouseEnter(): void {
            // Set timer to increase confidence that the user meant to interact with the
            // view, as opposed to accidentally moving past it. If the mouse leaves
            // the view quickly, clear the timeout for logging the event
            timeoutID = window.setTimeout(() => {
                props.telemetryService.log('InsightHover', { insightType: viewID.split('.')[0] })
            }, 500)

            viewContentElement?.addEventListener('mouseleave', onMouseLeave)
        }

        function onMouseLeave(): void {
            clearTimeout(timeoutID)
            viewContentElement?.removeEventListener('mouseleave', onMouseLeave)
        }

        // If containerClassName is specified, the element with this class is the element
        // that embodies the view in the eyes of the user. e.g. ViewGrid
        if (containerClassName) {
            viewContentElement = viewContentElement?.closest(`.${containerClassName}`) as HTMLDivElement
        }

        viewContentElement?.addEventListener('mouseenter', onMouseEnter)

        return () => {
            viewContentElement?.removeEventListener('mouseenter', onMouseEnter)
            viewContentElement?.removeEventListener('mouseleave', onMouseLeave)
            clearTimeout(timeoutID)
        }
    }, [viewID, containerClassName, props.telemetryService])

    return (
        <div className="view-content" ref={viewContentReference}>
            {viewContent.map((content, index) =>
                isMarkupContent(content) ? (
                    <React.Fragment key={index}>
                        {content.kind === MarkupKind.Markdown || !content.kind ? (
                            <Markdown
                                className="view-content__markdown mb-1"
                                dangerousInnerHTML={renderMarkdown(content.value)}
                                history={props.history}
                            />
                        ) : (
                            content.value
                        )}
                    </React.Fragment>
                ) : 'chart' in content ? (
                    <ChartViewContent key={index} content={content} viewID={viewID} history={props.history} />
                ) : content.component === 'QueryInput' ? (
                    <QueryInputInViewContent
                        {...props}
                        key={index}
                        implicitQueryPrefix={
                            typeof content.props.implicitQueryPrefix === 'string'
                                ? content.props.implicitQueryPrefix
                                : ''
                        }
                    />
                ) : null
            )}
        </div>
    )
}
