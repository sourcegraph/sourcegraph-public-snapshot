import React, { type FC, forwardRef, type HTMLAttributes, type PropsWithChildren } from 'react'

import { mdiRefresh } from '@mdi/js'
import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'

import {
    Button,
    type ForwardReferenceComponent,
    Icon,
    LegendItem,
    LegendList,
    type Series,
} from '@sourcegraph/wildcard'

import { InsightCard, InsightCardBanner, InsightCardHeader, InsightCardLoading } from '../../views'

import styles from './LivePreviewCard.module.scss'

interface LivePreviewCardProps extends HTMLAttributes<HTMLElement> {}

const LivePreviewCard: React.FunctionComponent<React.PropsWithChildren<LivePreviewCardProps>> = props => (
    <InsightCard {...props} className={classNames(styles.insightCard, props.className)} />
)

export interface LivePreviewUpdateButtonProps {
    disabled: boolean
    onClick: () => void
}

const LivePreviewUpdateButton: FC<LivePreviewUpdateButtonProps> = props => {
    const { disabled, onClick } = props

    return (
        <Button
            aria-label="Update code insight live preview"
            variant="icon"
            disabled={disabled}
            className={styles.updateButton}
            onClick={onClick}
        >
            Live preview
            <Icon svgPath={mdiRefresh} inline={false} aria-hidden={true} height="1rem" width="1rem" />
        </Button>
    )
}

const LivePreviewLoading: FC<PropsWithChildren<unknown>> = props => (
    <InsightCardLoading {...props} aria-label="Loading insight live preview" />
)

const LivePreviewHeader = InsightCardHeader

const LivePreviewBlurBackdrop = forwardRef((props, reference) => {
    const { as: Component = 'svg', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={reference}
            aria-hidden={true}
            className={classNames(styles.chartWithMock, className)}
        />
    )
}) as ForwardReferenceComponent<'svg', {}>

const LivePreviewBanner: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <InsightCardBanner className={styles.disableBanner}>{props.children}</InsightCardBanner>
)

interface LivePreviewChartProps extends React.ComponentProps<typeof ParentSize> {}

const LivePreviewChart: React.FunctionComponent<LivePreviewChartProps> = props => (
    <ParentSize {...props} className={classNames(styles.chartBlock, props.className)} />
)

interface LivePreviewLegendProps {
    series: Series<unknown>[]
}

const LivePreviewLegend: React.FunctionComponent<React.PropsWithChildren<LivePreviewLegendProps>> = props => {
    const { series } = props

    return (
        <LegendList aria-label="Live preview chart legend" className="mt-3">
            {series.map(series => (
                <LegendItem key={series.id} color={series.color} name={series.name} />
            ))}
        </LegendList>
    )
}

export {
    LivePreviewCard,
    LivePreviewHeader,
    LivePreviewUpdateButton,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewLegend,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
}
