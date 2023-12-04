import React, {
    type FC,
    forwardRef,
    type HTMLAttributes,
    type ReactNode,
    type PropsWithChildren,
    useState,
    type FocusEvent,
    createContext,
    useMemo,
    useContext,
} from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import {
    Card,
    type ForwardReferenceComponent,
    H2,
    H4,
    LoadingSpinner,
    LegendItem,
    LegendList,
    type Series,
} from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../../../../components/ErrorBoundary'

import styles from './InsightCard.module.scss'

interface CardContextData {
    isFocused: boolean
}

const CardContext = createContext<CardContextData>({ isFocused: false })

const InsightCard = forwardRef(function InsightCard(props, reference) {
    const { title, children, className, as = 'section', ...attributes } = props

    const [isFocused, setFocused] = useState(false)

    const handleFocusCapture = (): void => {
        if (!isFocused) {
            setFocused(true)
        }
    }

    const handleBlurCapture = (event: FocusEvent): void => {
        const nextFocusTarget = event.relatedTarget as Element | null

        if (!event.currentTarget.contains(nextFocusTarget)) {
            setFocused(false)
        }
    }

    return (
        <Card
            as={as}
            ref={reference}
            tabIndex={0}
            className={classNames(className, styles.view)}
            onFocusCapture={handleFocusCapture}
            onBlurCapture={handleBlurCapture}
            {...attributes}
        >
            <CardContext.Provider value={useMemo(() => ({ isFocused }), [isFocused])}>
                <ErrorBoundary location={useLocation()} className={styles.errorBoundary}>
                    {children}
                </ErrorBoundary>
            </CardContext.Provider>
        </Card>
    )
}) as ForwardReferenceComponent<'section'>

export interface InsightCardTitleProps {
    title: ReactNode
    subtitle?: ReactNode

    /**
     * It's primarily conceived as a slot for card actions (like filter buttons) that render
     * element right after header text at the right top corner of the card.
     */
    children?: ReactNode
}

const InsightCardHeader = forwardRef(function InsightCardHeader(props, reference) {
    const { as: Component = 'header', title, subtitle, className, children, ...attributes } = props

    return (
        <Component {...attributes} ref={reference} className={classNames(styles.header, className)}>
            <div className={styles.headerContent}>
                <H4
                    // We have to cast this element to H2 because having h4 without h3 and h2
                    // higher in the hierarchy violates a11y rules about headings structure.
                    as={H2}
                    className={styles.title}
                >
                    {title}
                </H4>

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

const InsightCardLoading: FC<PropsWithChildren<HTMLAttributes<HTMLElement>>> = props => {
    const { 'aria-label': ariaLabel = 'Insight loading', children, ...attributes } = props
    const { isFocused } = useContext(CardContext)

    return (
        <InsightCardBanner {...attributes}>
            <LoadingSpinner aria-label={ariaLabel} aria-live={isFocused ? 'polite' : 'off'} />
            {children}
        </InsightCardBanner>
    )
}

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
                <LegendItem key={series.id} color={series.color} name={series.name} />
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
