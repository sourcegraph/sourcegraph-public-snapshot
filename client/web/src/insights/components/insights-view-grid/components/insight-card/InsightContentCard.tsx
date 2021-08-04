import classnames from 'classnames'
import { MdiReactIconComponentType } from 'mdi-react'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { PropsWithChildren } from 'react'
import { useLocation } from 'react-router-dom'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import { ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { InsightTypePrefix } from '../../../../core/types'

import { InsightCardMenu } from './components/insight-card-menu/InsightCardMenu'
import styles from './InsightCard.module.scss'

const ASYNC_NOOP = (): Promise<void> => Promise.resolve()

export interface InsightCardProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    /**
     * Insight data (title, chart content)
     */
    insight: ViewProviderResult

    /**
     * Deleting handler fires when the user clicks delete in the insight menu.
     */
    onDelete?: (id: string) => void

    /**
     * Prop for enabling and disabling insight context menu.
     * Now only insight page has insights with context menu.
     */
    hasContextMenu?: boolean

    /**
     * To get container to track hovers for pings
     */
    containerClassName?: string
}

/**
 * Renders insight card content. Loading state, error state and insight itself.
 */
export const InsightContentCard: React.FunctionComponent<PropsWithChildren<InsightCardProps>> = props => {
    const {
        insight: { id, view },
        containerClassName,
        hasContextMenu,
        onDelete = ASYNC_NOOP,
        telemetryService,
        children,
        ...otherProps
    } = props

    const location = useLocation()

    // We support actions only over search and lang insights and not able to edit or delete
    // custom insight or backend insight.
    const hasMenu =
        hasContextMenu && (id.startsWith(InsightTypePrefix.search) || id.startsWith(InsightTypePrefix.langStats))

    const title = !isErrorLike(view) ? view?.title : null
    const subtitle = !isErrorLike(view) ? view?.subtitle : null

    // In case if we don't have a content for the header component
    // we should render nothing
    const hasHeader = title || subtitle || hasMenu

    return (
        <section
            {...otherProps}
            data-testid={`insight-card.${id}`}
            className={classnames('card', otherProps.className, styles.insightCard)}
        >
            <ErrorBoundary
                className="pt-0"
                location={location}
                extraContext={
                    <>
                        <p>ID: {id}</p>
                        <pre>View: {JSON.stringify(view, null, 2)}</pre>
                    </>
                }
            >
                {hasHeader && (
                    <header className={styles.insightCardHeader}>
                        <div className={styles.insightCardHeaderContent}>
                            <h4 className={styles.insightCardTitle}>{title}</h4>
                            {subtitle && <div className={styles.insightCardSubtitle}>{subtitle}</div>}
                        </div>

                        {hasMenu && (
                            <InsightCardMenu
                                menuButtonClassName="mr-n2 d-inline-flex"
                                insightID={id}
                                onDelete={onDelete}
                            />
                        )}
                    </header>
                )}

                {children}
            </ErrorBoundary>
        </section>
    )
}

export const getInsightViewIcon = (source: ViewInsightProviderSourceType): MdiReactIconComponentType => {
    switch (source) {
        case ViewInsightProviderSourceType.Backend:
            return DatabaseIcon
        case ViewInsightProviderSourceType.Extension:
            return PuzzleIcon
    }
}
