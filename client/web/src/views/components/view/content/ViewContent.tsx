import classNames from 'classnames'
import { isObject } from 'lodash'
import React, { useCallback, useEffect, useRef } from 'react'
import { View, MarkupContent } from 'sourcegraph'

import { ErrorLike } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { hasProperty } from '@sourcegraph/shared/src/util/types'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../components/alerts'

import { ChartViewContent, ChartViewContentLayout } from './chart-view-content/ChartViewContent'
import styles from './ViewContent.module.scss'

const isMarkupContent = (input: unknown): input is MarkupContent =>
    isObject(input) && hasProperty('value')(input) && typeof input.value === 'string'

export interface ViewContentProps extends TelemetryProps {
    content: View['content']

    /**
     * View tracking type is used to send a proper pings event (InsightHover, InsightDataPointClick)
     * with view type as a tracking variable. If it equals to null tracking is disabled.
     *
     * Note: these events are code insights specific, we have to remove this tracking logic to
     * reuse this component in other consumers.
     */
    viewTrackingType?: string

    /** To get container to track hovers for pings */
    containerClassName?: string

    /** Optionally display an alert overlay */
    alert?: React.ReactNode
    layout?: ChartViewContentLayout
    className?: string
}

/**
 * Renders the content sections of the view. It supports markdown and different types
 * of chart views.
 *
 * Support for non-MarkupContent elements is experimental and subject to change or removal
 * without notice.
 */
export const ViewContent: React.FunctionComponent<ViewContentProps> = props => {
    const { content, viewTrackingType, containerClassName, alert, layout, className, telemetryService } = props

    // Track user intent to interact with extension-contributed views
    const viewContentReference = useRef<HTMLDivElement>(null)

    // TODO Move this tracking logic out of this shared view component
    useEffect(() => {
        // Disable tracking logic if view tracking type wasn't provided
        if (!viewTrackingType) {
            return
        }

        let viewContentElement = viewContentReference.current
        let timeoutID: number | undefined

        function onMouseEnter(): void {
            // Set timer to increase confidence that the user meant to interact with the
            // view, as opposed to accidentally moving past it. If the mouse leaves
            // the view quickly, clear the timeout for logging the event
            // TODO: Move this from common component to shared logic/hook
            timeoutID = window.setTimeout(() => {
                telemetryService.log(
                    'InsightHover',
                    { insightType: viewTrackingType },
                    { insightType: viewTrackingType }
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
    }, [viewTrackingType, containerClassName, telemetryService])

    const handleOnDatumLinkClick = useCallback(() => {
        // Disable tracking logic if view tracking type wasn't provided
        if (!viewTrackingType) {
            return
        }

        // TODO: Move this from common component to shared logic/hook
        telemetryService.log(
            'InsightDataPointClick',
            { insightType: viewTrackingType },
            { insightType: viewTrackingType }
        )
    }, [viewTrackingType, telemetryService])

    return (
        <div className={classNames(styles.viewContent, className)} ref={viewContentReference}>
            {content.map((content, index) =>
                isMarkupContent(content) ? (
                    <React.Fragment key={index}>
                        {content.kind === MarkupKind.Markdown || !content.kind ? (
                            <Markdown
                                className={classNames('mb-1', styles.markdown)}
                                dangerousInnerHTML={renderMarkdown(content.value)}
                            />
                        ) : (
                            content.value
                        )}
                    </React.Fragment>
                ) : 'chart' in content ? (
                    <React.Fragment key={index}>
                        {alert && <div className={styles.viewContentAlertOverlay}>{alert}</div>}
                        <ChartViewContent
                            content={content}
                            layout={layout}
                            className="flex-grow-1"
                            onDatumLinkClick={handleOnDatumLinkClick}
                        />
                    </React.Fragment>
                ) : null
            )}
        </div>
    )
}

export interface ViewErrorContentProps {
    title: string
    error: ErrorLike
}

export const ViewErrorContent: React.FunctionComponent<ViewErrorContentProps> = props => {
    const { error, title, children } = props

    return (
        <div className="h-100 w-100 d-flex flex-column">
            {children || <ErrorAlert data-testid={`${title} view error`} className="m-0" error={error} />}
        </div>
    )
}

export interface ViewLoadingContentProps {
    text: string
}

export const ViewLoadingContent: React.FunctionComponent<ViewLoadingContentProps> = props => {
    const { text } = props

    return (
        <div className="h-100 w-100 d-flex flex-column flex-grow-1">
            <span className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                <LoadingSpinner /> {text}
            </span>
        </div>
    )
}

export const ViewBannerContent: React.FunctionComponent = ({ children }) => (
    <div className="h-100 w-100 d-flex flex-column justify-content-center align-items-center">{children}</div>
)
