import { isObject } from 'lodash'
import React, { useEffect, useRef, useState } from 'react'
import { View, MarkupContent } from 'sourcegraph'

import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { hasProperty } from '@sourcegraph/shared/src/util/types'

import { ChartViewContent } from './chart-view-content/ChartViewContent'

const isMarkupContent = (input: unknown): input is MarkupContent =>
    isObject(input) && hasProperty('value')(input) && typeof input.value === 'string'

export interface InsightViewContentProps extends TelemetryProps {
    viewContent: View['content']
    viewID: string

    /** To get container to track hovers for pings */
    containerClassName?: string

    /** Render prop to display overlaid alerts */
    alertOverlay?(): void
}

/**
 * Renders the content of an extension-contributed view.
 */
export const InsightViewContent: React.FunctionComponent<InsightViewContentProps> = ({
    viewContent,
    viewID,
    containerClassName,
    alertOverlay,
    ...props
}) => {
    // Track hovering for hide/show message overlay
    const [isHovered, setIsHovered] = useState(false)
    const hoverProps = {
        onMouseOver: () => setIsHovered(true),
        onMouseOut: () => setIsHovered(false),
        onFocus: () => setIsHovered(true),
        onBlur: () => setIsHovered(false),
    }

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
                props.telemetryService.log(
                    'InsightHover',
                    { insightType: viewID.split('.')[0] },
                    { insightType: viewID.split('.')[0] }
                )
            }, 500)

            viewContentElement?.addEventListener('mouseleave', onMouseLeave)
        }

        function onMouseLeave(): void {
            clearTimeout(timeoutID)
            viewContentElement?.removeEventListener('mouseleave', onMouseLeave)
        }

        // If containerClassName is specified, the element with this class is the element
        // that embodies the view in the eyes of the user. e.g. InsightsViewGrid
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
        <div className="view-content" ref={viewContentReference} {...hoverProps}>
            {viewContent.map((content, index) =>
                isMarkupContent(content) ? (
                    <React.Fragment key={index}>
                        {content.kind === MarkupKind.Markdown || !content.kind ? (
                            <Markdown
                                className="view-content__markdown mb-1"
                                dangerousInnerHTML={renderMarkdown(content.value)}
                            />
                        ) : (
                            content.value
                        )}
                    </React.Fragment>
                ) : 'chart' in content ? (
                    <React.Fragment key={index}>
                        <ChartViewContent
                            content={content}
                            viewID={viewID}
                            telemetryService={props.telemetryService}
                            className="view-content__chart"
                        />
                        {alertOverlay && !isHovered && alertOverlay()}
                    </React.Fragment>
                ) : null
            )}
        </div>
    )
}
