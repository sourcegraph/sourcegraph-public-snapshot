import React from 'react'

import classNames from 'classnames'
import { isObject } from 'lodash'
import { View, MarkupContent } from 'sourcegraph'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, hasProperty, renderMarkdown } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ChartViewContent, ChartViewContentLayout } from './chart-view-content/ChartViewContent'

import styles from './ViewContent.module.scss'

const isMarkupContent = (input: unknown): input is MarkupContent =>
    isObject(input) && hasProperty('value')(input) && typeof input.value === 'string'

export interface ViewContentProps extends React.HTMLAttributes<HTMLElement> {
    content: View['content']
    alert?: React.ReactNode
    layout?: ChartViewContentLayout
    className?: string

    /**
     * Calls whenever the user clicks on or navigate through
     * the chart datum (pie arc, line point, bar category)
     */
    onDatumLinkClick?: () => void
    locked?: boolean
}

/**
 * Renders the content sections of the view. It supports markdown and different types
 * of chart views.
 *
 * Support for non-MarkupContent elements is experimental and subject to change or removal
 * without notice.
 */
export const ViewContent: React.FunctionComponent<React.PropsWithChildren<ViewContentProps>> = props => {
    const { content, alert, layout, className, onDatumLinkClick, locked = false, ...attributes } = props

    return (
        <div {...attributes} className={classNames(styles.viewContent, className)}>
            {content.map((content, index) =>
                // TODO Remove markup content support
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
                            onDatumLinkClick={onDatumLinkClick}
                            locked={locked}
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

export const ViewErrorContent: React.FunctionComponent<React.PropsWithChildren<ViewErrorContentProps>> = props => {
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

export const ViewLoadingContent: React.FunctionComponent<React.PropsWithChildren<ViewLoadingContentProps>> = props => {
    const { text } = props

    return (
        <div className="h-100 w-100 d-flex flex-column flex-grow-1">
            <span className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                <LoadingSpinner /> {text}
            </span>
        </div>
    )
}

export const ViewBannerContent: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <div className="h-100 w-100 d-flex flex-column justify-content-center align-items-center">{children}</div>
)
