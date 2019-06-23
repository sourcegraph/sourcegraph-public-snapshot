import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { ThreadDiscussionPage } from '../../threads/detail/discussion/ThreadDiscussionPage'
import { ThreadSettingsPage } from '../../threads/detail/settings/ThreadSettingsPage'
import { ThreadAreaSidebar } from '../../threads/detail/sidebar/ThreadAreaSidebar'
import { createThreadAreaContext, ThreadAreaContext } from '../../threads/detail/ThreadArea'
import { useChangesetByID } from '../components/useChangesetByID'
import { ChangesetsAreaContext } from '../global/ChangesetsArea'
import { ChangesetChangesPage } from './changes/ChangesetChangesPage'
import { ChangesetAreaNavbar } from './navbar/ChangesetAreaNavbar'
import { ChangesetOverview } from './overview/ChangesetOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends ChangesetsAreaContext, RouteComponentProps<{ threadID: string }> {}

export interface ChangesetAreaContext extends ThreadAreaContext {}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single changeset.
 */
export const ChangesetArea: React.FunctionComponent<Props> = props => {
    const [threadOrError, setThreadOrError] = useChangesetByID(props.match.params.threadID)
    if (threadOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(threadOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={threadOrError.message} />
    }
    const context = createThreadAreaContext(props, { thread: threadOrError, onThreadUpdate: setThreadOrError })
    return (
        <div className="changeset-area flex-1 d-flex overflow-hidden">
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <ChangesetOverview
                        {...context}
                        location={props.location}
                        history={props.history}
                        className="container flex-0 pb-3"
                    />
                    <div className="w-100 border-bottom" />
                    <ChangesetAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route
                            path={props.match.url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadDiscussionPage
                                    {...context}
                                    {...routeComponentProps}
                                    className="container mb-3"
                                />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/changes`}
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ChangesetChangesPage {...context} {...routeComponentProps} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/settings`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadSettingsPage {...context} {...routeComponentProps} className="container mt-3" />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
            <ThreadAreaSidebar {...context} className="changeset-area__sidebar flex-0" history={props.history} />
        </div>
    )
}
