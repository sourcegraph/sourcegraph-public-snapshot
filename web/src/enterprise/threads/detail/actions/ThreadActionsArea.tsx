import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorBoundary } from '../../../../components/ErrorBoundary'
import { HeroPage } from '../../../../components/HeroPage'
import { ThreadSettings } from '../../settings'
import { ThreadAreaContext } from '../ThreadArea'
import { ThreadActionsCommitStatusesPage } from './commitStatuses/ThreadActionsCommitStatusesPage'
import { ThreadActionsEmailNotificationsPage } from './email/ThreadActionsEmailNotificationsPage'
import { ThreadActionsPullRequestsPage } from './pullRequests/ThreadActionsPullRequestsPage'
import { ThreadActionsSlackNotificationsPage } from './slackNotifications/ThreadActionsSlackNotificationsPage'
import { ThreadActionsAreaSidebar } from './ThreadActionsAreaSidebar'
import { ThreadActionsOverview } from './ThreadActionsOverview'
import { ThreadActionsWebhooksPage } from './webhooks/ThreadActionsWebhooksPage'
import { ThreadActionsEditorPage } from './editor/ThreadActionsEditorPage'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends ExtensionsControllerProps, RouteComponentProps<{}> {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The actions area for a single thread.
 */
export const ThreadActionsArea: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    className = '',
    ...props
}) => {
    const context: ThreadAreaContext & { areaURL: string } & ExtensionsControllerProps = {
        thread,
        onThreadUpdate,
        threadSettings,
        areaURL: props.match.url,
        extensionsController: props.extensionsController,
    }

    return (
        <div className={`thread-actions-area d-flex ${className} mb-3`}>
            <ThreadActionsAreaSidebar {...context} className="flex-0 mr-3" />
            <div className="flex-1">
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route
                            path={props.match.url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={_routeComponentProps => <ThreadActionsOverview />}
                        />
                        <Route
                            path={`${props.match.url}/pull-requests`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsPullRequestsPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/commit-statuses`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsCommitStatusesPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/slack`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsSlackNotificationsPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/email`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsEmailNotificationsPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/editor`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsEditorPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route
                            path={`${props.match.url}/webhooks`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsWebhooksPage {...routeComponentProps} {...context} />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
