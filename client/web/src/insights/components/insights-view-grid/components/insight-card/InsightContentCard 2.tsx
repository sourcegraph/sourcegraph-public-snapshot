import classnames from 'classnames'
import { noop } from 'lodash'
import { MdiReactIconComponentType } from 'mdi-react'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { PropsWithChildren, ReactNode, useState } from 'react'
import { useLocation } from 'react-router-dom'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import { ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { InsightTypePrefix } from '../../../../core/types'
import { LineChartSettingsContext } from '../../../insight-view-content/chart-view-content/charts/line/line-chart-settings-provider'

import { InsightCardMenu } from './components/insight-card-menu/InsightCardMenu'
import styles from './InsightCard.module.scss'

export interface InsightCardProps
    extends TelemetryProps,
        React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    /**
     * Insight data (title, chart content)
     */
    insight: ViewProviderResult

    /**
     * Custom card actions (like filter buttons) that render element right next to three dots
     * card context menu.
     */
    actions?: ReactNode

    /**
     * Prop for enabling and disabling insight context menu.
     * Now only insight page has insights with context menu.
     */
    hasContextMenu?: boolean

    /**
     * To get container to track hovers for pings
     */
    containerClassName?: string

    /**
     * Ref prop for root element (section) of insight content card.
     */
    innerRef?: React.RefObject<HTMLElement>

    /**
     * Deleting handler fires when the user clicks delete in the insight menu.
     */
    onDelete?: (id: string) => void
}

/**
 * Renders insight card content. Loading state, error state and insight itself.
 */
export const InsightContentCard: React.FunctionComponent<PropsWithChildren<InsightCardProps>> = props => {
    const {
        insight: { id, view },
        actions,
        containerClassName,
        hasContextMenu,
        telemetryService,
        children,
        innerRef,
        onDelete = noop,
        ...otherProps
    } = props

    const location = useLocation()

    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const handleToggleZeroYAxisMin = (): void => setZeroYAxisMin(!zeroYAxisMin)

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
        <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
            <section
                {...otherProps}
                data-testid={`insight-card.${id}`}
                /* eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex */
                tabIndex={0}
                ref={innerRef}
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

                            <div className="align-self-start d-flex align-items-center">
                                {actions}
                                {hasMenu && (
                                    <InsightCardMenu
                                        menuButtonClassName="ml-1 mr-n2 d-inline-flex"
                                        insightID={id}
                                        onDelete={onDelete}
                                        onToggleZeroYAxisMin={handleToggleZeroYAxisMin}
                                    />
                                )}
                            </div>
                        </header>
                    )}

                    {children}
                </ErrorBoundary>
            </section>
        </LineChartSettingsContext.Provider>
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
