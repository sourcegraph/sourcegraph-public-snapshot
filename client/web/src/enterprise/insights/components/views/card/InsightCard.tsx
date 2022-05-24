import React, { forwardRef, HTMLAttributes, ReactNode } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { Card, ForwardReferenceComponent, Typography, LoadingSpinner } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, Series } from '../../../../../charts'
import { ErrorBoundary } from '../../../../../components/ErrorBoundary'

import styles from './InsightCard.module.scss'

const InsightCard = forwardRef((props, reference) => {
    const { title, children, className, as = 'section', ...otherProps } = props

    return (
        <Card as={as} tabIndex={0} ref={reference} {...otherProps} className={classNames(className, styles.view)}>
            <ErrorBoundary location={useLocation()} className={styles.errorBoundary}>
                {children}
            </ErrorBoundary>
        </Card>
    )
}) as ForwardReferenceComponent<'section'>

export interface InsightCardTitleProps {
    title: string
    subtitle?: string

    /**
     * It's primarily conceived as a slot for card actions (like filter buttons) that render
     * element right after header text at the right top corner of the card.
     */
    children?: ReactNode
}

const InsightCardHeader = forwardRef((props, reference) => {
    const { as: Component = 'header', title, subtitle, className, children, ...attributes } = props

    return (
        <Component {...attributes} ref={reference} className={classNames(styles.header, className)}>
            <div className={styles.headerContent}>
                <Typography.H4 as={Typography.H2} title={title} className={styles.title}>
                    {title}
                </Typography.H4>

                {children && (
                    // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
                    <div
                        // View component usually is rendered within a view grid component. To suppress
                        // bad click that lead to card DnD events in view grid we stop event bubbling for
                        // clicks.
                        onClick={event => event.stopPropagation()}
                        onMouseDown={event => event.stopPropagation()}
                        className={styles.action}
                    >
                        {children}
                    </div>
                )}
            </div>

            {subtitle && <div className={styles.subtitle}>{subtitle}</div>}
        </Component>
    )
}) as ForwardReferenceComponent<'header', InsightCardTitleProps>

const InsightCardLoading: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <InsightCardBanner>
        <LoadingSpinner />
        {props.children}
    </InsightCardBanner>
)

const InsightCardBanner: React.FunctionComponent<React.PropsWithChildren<HTMLAttributes<HTMLDivElement>>> = props => (
    <div {...props} className={classNames(styles.loadingContent, props.className)}>
        {props.children}
    </div>
)

interface InsightCardLegendProps extends React.HTMLAttributes<HTMLUListElement> {
    series: Series<any>[]
}

const InsightCardLegend: React.FunctionComponent<React.PropsWithChildren<InsightCardLegendProps>> = props => {
    const { series, ...attributes } = props

    return (
        <LegendList {...attributes}>
            {series.map(series => (
                <LegendItem key={series.id} color={getLineColor(series)} name={series.name} />
            ))}
        </LegendList>
    )
}

const Root = InsightCard
const Header = InsightCardHeader
const Loading = InsightCardLoading
const Banner = InsightCardBanner
const Legends = InsightCardLegend

export {
    InsightCard,
    InsightCardHeader,
    InsightCardLoading,
    InsightCardBanner,
    InsightCardLegend,
    // * as Card imports
    Root,
    Header,
    Loading,
    Banner,
    Legends,
}
