import React, { useMemo } from 'react'
import { isErrorLike } from '../../../../shared/src/util/errors'
import classNames from 'classnames'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '../../components/alerts'
import { ViewContent, ViewContentProps } from '../../views/ViewContent'
import * as H from 'history'
import { WidthProvider, Responsive, Layout as ReactGridLayout, Layouts as ReactGridLayouts } from 'react-grid-layout'
import { ViewProviderResult } from '../../../../shared/src/api/client/services/viewService'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { usePersistedGridLayouts } from '../../components/gridLayout/usePersistedGridLayouts'

// TODO use a method to get width that also triggers when file explorer is closed
// (WidthProvider only listens to window resize events)
const ResponsiveGridLayout = WidthProvider(Responsive)

export interface ViewGridProps extends Omit<ViewContentProps, 'viewContent'> {
    viewGridStorageKey: string
    views: ViewProviderResult[]
    className?: string
    history: H.History
}

const breakpointNames = ['xs', 'sm', 'md', 'lg'] as const
type BreakpointName = typeof breakpointNames[number]

/** Minimum size in px after which a breakpoint is active. */
const breakpoints: Record<BreakpointName, number> = { xs: 0, sm: 576, md: 768, lg: 992 } // no xl because TreePage's max-width is the xl breakpoint.
const columns: Record<BreakpointName, number> = { xs: 1, sm: 6, md: 8, lg: 12 }
const defaultItemsPerRow: Record<BreakpointName, number> = { xs: 1, sm: 2, md: 2, lg: 3 }
const minWidths: Record<BreakpointName, number> = { xs: 1, sm: 2, md: 3, lg: 3 }
const defaultHeight = 3

const viewsToReactGridLayouts = (views: ViewProviderResult[]): ReactGridLayouts => {
    const reactGridLayouts = Object.fromEntries(
        breakpointNames.map(
            breakpointName =>
                [
                    breakpointName,
                    views.map(
                        ({ id }, index): ReactGridLayout => {
                            const width = columns[breakpointName] / defaultItemsPerRow[breakpointName]
                            if (id === 'treeView.readme') {
                                return {
                                    i: id,
                                    h: 5 * defaultHeight,
                                    w: columns[breakpointName],
                                    x: (index * width) % columns[breakpointName],
                                    y: Math.floor((index * width) / columns[breakpointName]),
                                    minW: minWidths[breakpointName],
                                    minH: 2,
                                }
                            }
                            return {
                                i: id,
                                h: defaultHeight,
                                w: width,
                                x: (index * width) % columns[breakpointName],
                                y: Math.floor((index * width) / columns[breakpointName]),
                                minW: minWidths[breakpointName],
                                minH: 1,
                            }
                        }
                    ),
                ] as const
        )
    )
    return reactGridLayouts
}

export const ViewGrid: React.FunctionComponent<ViewGridProps> = props => {
    const allDefaultLayouts = useMemo(() => viewsToReactGridLayouts(props.views), [props.views])
    const [layouts, onLayoutChange] = usePersistedGridLayouts(props.viewGridStorageKey, allDefaultLayouts)

    return (
        <div className={classNames(props.className, 'view-grid')}>
            <ResponsiveGridLayout
                breakpoints={breakpoints}
                layouts={layouts}
                onLayoutChange={onLayoutChange}
                cols={columns}
                autoSize={true}
                rowHeight={6 * 16}
                containerPadding={[0, 0]}
                margin={[12, 12]}
            >
                {props.views.map(({ id, view }) => (
                    <div key={id} className={classNames('card view-grid__item')}>
                        {view === undefined ? (
                            <div className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                                <LoadingSpinner /> Loading
                            </div>
                        ) : isErrorLike(view) ? (
                            <ErrorAlert className="m-0" error={view} history={props.history} />
                        ) : (
                            <>
                                {view.title && (
                                    <h3 className="view-grid__view-title">
                                        <LinkOrSpan to={view.titleLink} style={{ color: 'var(--body-color)' }}>
                                            {view.title}
                                        </LinkOrSpan>
                                    </h3>
                                )}
                                {view.subtitle && <div className="view-grid__view-subtitle">{view.subtitle}</div>}
                                <ViewContent
                                    {...props}
                                    settingsCascade={props.settingsCascade}
                                    viewContent={view.content}
                                />
                            </>
                        )}
                    </div>
                ))}
            </ResponsiveGridLayout>
        </div>
    )
}
