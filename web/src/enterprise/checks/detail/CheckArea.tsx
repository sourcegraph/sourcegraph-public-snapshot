import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { ChecksAreaContext } from '../../status/statusesArea/ChecksArea'
import { Status } from '../status'
import { useCheckByTypeForScope } from '../util/useCheckByTypeForScope'
import { StatusChecksPage } from './checks/CheckChecksPage'
import { StatusIssuesPage } from './issues/CheckIssuesPage'
import { CheckAreaNavbar } from './navbar/CheckAreaNavbar'
import { StatusNotificationsPage } from './notifications/StatusNotificationsPage'
import { StatusOverview } from './overview/StatusOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends Pick<CheckAreaContext, Exclude<keyof CheckAreaContext, 'status'>> {}

export interface CheckAreaContext extends ChecksAreaContext, ExtensionsControllerProps, PlatformContextProps {
    /** The status name. */
    name: string

    /** The status. */
    status: Status

    /** The URL to the check area for this check. */
    statusURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single check.
 */
export const CheckArea: React.FunctionComponent<Props> = ({ name, scope, statusURL, ...props }) => {
    const checkOrError = useCheckByTypeForScope(props.extensionsController, name, scope)
    if (checkOrError === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (checkOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Check not found" />
    }
    if (isErrorLike(checkOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={checkOrError.message} />
    }

    const context: CheckAreaContext = {
        ...props,
        name,
        scope,
        status: checkOrError,
        statusURL,
    }

    return (
        <div className="status-area flex-1 d-flex overflow-hidden">
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <StatusOverview {...context} className="container flex-0 pb-3" />
                    <div className="w-100 border-bottom" />
                    <CheckAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route path={statusURL} exact={true}>
                            <StatusNotificationsPage {...context} className="mt-3 container" />
                        </Route>
                        <Route path={`${statusURL}/checks`}>
                            <StatusChecksPage {...context} />
                        </Route>
                        <Route path={`${statusURL}/issues`} exact={true}>
                            <StatusIssuesPage {...context} />
                        </Route>
                        <Route component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
