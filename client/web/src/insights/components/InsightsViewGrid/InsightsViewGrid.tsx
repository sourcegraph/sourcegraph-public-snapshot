import classNames from 'classnames'
import { MdiReactIconComponentType } from 'mdi-react'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useCallback, useMemo } from 'react'
import { Layout as ReactGridLayout, Layouts as ReactGridLayouts, Responsive, WidthProvider } from 'react-grid-layout'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { ViewContent, ViewContentProps } from '../../../views/ViewContent'
import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../../core/backend/types'

// TODO use a method to get width that also triggers when file explorer is closed
// (WidthProvider only listens to window resize events)
const ResponsiveGridLayout = WidthProvider(Responsive)

export interface InsightsViewGridProps
    extends Omit<ViewContentProps, 'viewContent' | 'viewID' | 'containerClassName'>,
        TelemetryProps {
    views: ViewInsightProviderResult[]
    className?: string
}

const breakpointNames = ['xs', 'sm', 'md', 'lg'] as const
type BreakpointName = typeof breakpointNames[number]

/** Minimum size in px after which a breakpoint is active. */
const breakpoints: Record<BreakpointName, number> = { xs: 0, sm: 576, md: 768, lg: 992 } // no xl because TreePage's max-width is the xl breakpoint.
const columns: Record<BreakpointName, number> = { xs: 1, sm: 6, md: 8, lg: 12 }
const defaultItemsPerRow: Record<BreakpointName, number> = { xs: 1, sm: 2, md: 2, lg: 3 }
const minWidths: Record<BreakpointName, number> = { xs: 1, sm: 2, md: 3, lg: 3 }
const defaultHeight = 3

const viewsToReactGridLayouts = (views: ViewInsightProviderResult[]): ReactGridLayouts => {
    const reactGridLayouts = Object.fromEntries(
        breakpointNames.map(
            breakpointName =>
                [
                    breakpointName,
                    views.map(
                        ({ id }, index): ReactGridLayout => {
                            const width = columns[breakpointName] / defaultItemsPerRow[breakpointName]
                            return {
                                i: id,
                                h: defaultHeight,
                                w: width,
                                x: (index * width) % columns[breakpointName],
                                y: Math.floor((index * width) / columns[breakpointName]),
                                minW: minWidths[breakpointName],
                                minH: 2,
                            }
                        }
                    ),
                ] as const
        )
    )
    return reactGridLayouts
}

interface InsightDescriptionProps {
    title: string
    icon: MdiReactIconComponentType
    className?: string
}

// Since we use react-grid-layout for build draggable insight cards at insight dashboard
// to support text selection within insight card at InsightDescription component we have to
// capture mouse event to prevent all action from react-grid-layout library which will prevent
// default behavior and the text will become unavailable for selection
const stopPropagation: React.MouseEventHandler<HTMLElement> = event => event.stopPropagation()

const InsightDescription: React.FunctionComponent<InsightDescriptionProps> = props => {
    const { icon: Icon, title, className = '' } = props

    return (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <small
            title={title}
            className={classNames('insight-description', 'text-muted', className)}
            onMouseDown={stopPropagation}
        >
            <Icon className="icon-inline" /> {title}
        </small>
    )
}

const getInsightViewIcon = (source: ViewInsightProviderSourceType): MdiReactIconComponentType => {
    switch (source) {
        case ViewInsightProviderSourceType.Backend:
            return DatabaseIcon
        case ViewInsightProviderSourceType.Extension:
            return PuzzleIcon
    }
}

export const InsightsViewGrid: React.FunctionComponent<InsightsViewGridProps> = props => {
    const onResizeOrDragStart: ReactGridLayout.ItemCallback = useCallback(
        (_layout, item) => {
            try {
                props.telemetryService.log('InsightUICustomization', { insightType: item.i.split('.')[0] })
            } catch {
                // noop
            }
        },
        [props.telemetryService]
    )

    // For Firefox we can't use css transform/translate to put view grid item in right place.
    // Instead of this we have to fallback on css position:absolute top/left properties.
    // Reason: Since we use coordinate transformation svg.getScreenCTM from svg viewport
    // to dom viewport in order to calculate chart tooltip position and firefox has
    // a bug about providing right value of screenCTM matrix when parent has css transform.
    // https://bugzilla.mozilla.org/show_bug.cgi?id=1610093
    // Back to css transforms when this bug will be resolved in Firefox.
    const useCSSTransforms = useMemo(() => !isFirefox(), [])

    return (
        <div className={classNames(props.className, 'insights-view-grid')}>
            <ResponsiveGridLayout
                breakpoints={breakpoints}
                layouts={viewsToReactGridLayouts(props.views)}
                cols={columns}
                autoSize={true}
                rowHeight={6 * 16}
                containerPadding={[0, 0]}
                useCSSTransforms={useCSSTransforms}
                margin={[12, 12]}
                onResizeStart={onResizeOrDragStart}
                onDragStart={onResizeOrDragStart}
            >
                {props.views.map(({ id, view, source }) => (
                    <div key={id} className={classNames('card insights-view-grid__item')}>
                        <ErrorBoundary
                            location={props.location}
                            extraContext={
                                <>
                                    <p>ID: {id}</p>
                                    <pre>View: {JSON.stringify(view, null, 2)}</pre>
                                </>
                            }
                            className="pt-0"
                        >
                            {view === undefined ? (
                                <>
                                    <div className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                                        <LoadingSpinner /> Loading code insight
                                    </div>
                                    <InsightDescription
                                        className="insights-view-grid__view-description"
                                        title={id}
                                        icon={getInsightViewIcon(source)}
                                    />
                                </>
                            ) : isErrorLike(view) ? (
                                <>
                                    <ErrorAlert className="m-0" error={view} />
                                    <InsightDescription
                                        className="insights-view-grid__view-description"
                                        title={id}
                                        icon={getInsightViewIcon(source)}
                                    />
                                </>
                            ) : (
                                <>
                                    <h3 className="insights-view-grid__view-title">{view.title}</h3>
                                    {view.subtitle && (
                                        <div className="insights-view-grid__view-subtitle">{view.subtitle}</div>
                                    )}
                                    <ViewContent
                                        {...props}
                                        settingsCascade={props.settingsCascade}
                                        viewContent={view.content}
                                        viewID={id}
                                        containerClassName="insights-view-grid__item"
                                    />
                                </>
                            )}
                        </ErrorBoundary>
                    </div>
                ))}
            </ResponsiveGridLayout>
        </div>
    )
}
