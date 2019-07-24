import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, Switch } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { parseJSON } from '../../settings/configuration'
import { RouteDescriptor } from '../../util/contributions'
import { ThreadSettings } from '../threads/settings'
import { ThreadlikeAreaNavbar } from './navbar/ThreadlikeAreaNavbar'
import { ThreadlikeAreaSidebar } from './sidebar/ThreadlikeAreaSidebar'

interface ThreadlikeProps {
    /** The thread. */
    thread: GQL.IThread

    /** Called to update the thread. */
    onThreadUpdate: (thread: GQL.IThread) => void
}

export interface ThreadlikeAreaContext extends ThreadlikeProps {
    /** The thread's parsed JSON settings. */
    threadSettings: ThreadSettings

    location: H.Location
    history: H.History
}

export interface ThreadlikePage extends RouteDescriptor<ThreadlikeAreaContext> {
    title: string
    icon?: React.ComponentType<{ className?: string }>
    count?: number
}

interface Props extends ThreadlikeProps {
    /**
     * The React component that renders the overview, which is shown above the page tab bar.
     */
    overviewComponent: React.ComponentType<ThreadlikeAreaContext & { className?: string }>

    /**
     * The pages to show for this threadlike.
     */
    pages: ThreadlikePage[]

    className?: string
    location: H.Location
    history: H.History
}

/**
 * The area for a single threadlike.
 */
export const ThreadlikeArea: React.FunctionComponent<Props> = ({
    overviewComponent: OverviewComponent,
    pages: conditionalPages,
    className = '',
    ...props
}) => {
    const context: ThreadlikeAreaContext = {
        ...props,
        threadSettings: parseJSON(props.thread.settings),
    }
    const pages = conditionalPages.filter(
        (page): page is Pick<typeof page, Exclude<keyof typeof page, 'condition'>> =>
            !page.condition || page.condition(context)
    )
    return (
        <div className={`threadlike-area flex-1 d-flex overflow-hidden ${className}`}>
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <OverviewComponent {...context} className="container flex-0 pb-3" />
                    <div className="w-100 border-bottom" />
                    <ThreadlikeAreaNavbar {...context} pages={pages} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        {pages.map((page, i) => (
                            <Route
                                key={i}
                                path={page.path}
                                exact={page.exact}
                                // tslint:disable-next-line: jsx-no-lambda
                                render={routeComponentProps => page.render({ ...routeComponentProps, ...context })}
                            />
                        ))}
                        <Route>
                            <HeroPage
                                icon={MapSearchIcon}
                                title="404: Not Found"
                                subtitle="Sorry, the requested page was not found."
                            />
                        </Route>
                    </Switch>
                </ErrorBoundary>
            </div>
            <ThreadlikeAreaSidebar {...context} className="threadlike-area__sidebar flex-0" history={props.history} />
        </div>
    )
}
