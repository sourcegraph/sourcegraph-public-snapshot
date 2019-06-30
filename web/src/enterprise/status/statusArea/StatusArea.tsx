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
import { Status } from '../status'
import { StatusesAreaContext } from '../statusesArea/StatusesArea'
import { useStatusByTypeForScope } from '../util/useStatusByTypeForScope'
import { StatusAreaNavbar } from './navbar/StatusAreaNavbar'
import { StatusOverview } from './overview/StatusOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends Pick<StatusAreaContext, Exclude<keyof StatusAreaContext, 'status'>> {}

export interface StatusAreaContext extends StatusesAreaContext, ExtensionsControllerProps, PlatformContextProps {
    /** The status name. */
    name: string

    /** The status. */
    status: Status

    /** The URL to the status area for this status. */
    statusURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single status.
 */
export const StatusArea: React.FunctionComponent<Props> = ({ name, scope, statusURL, ...props }) => {
    const statusOrError = useStatusByTypeForScope(props.extensionsController, name, scope)
    if (statusOrError === LOADING) {
        return null // loading
    }
    if (statusOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Status not found" />
    }
    if (isErrorLike(statusOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={statusOrError.message} />
    }

    const context: StatusAreaContext = {
        ...props,
        name,
        scope,
        status: statusOrError,
        statusURL,
    }

    return (
        <div className="status-area flex-1 d-flex overflow-hidden">
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <StatusOverview
                        {...context}
                        location={props.location}
                        history={props.history}
                        className="container flex-0 pb-3"
                    />
                    <div className="w-100 border-bottom" />
                    <StatusAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route path={statusURL} exact={true}>
                            <p>hello, world!</p>
                        </Route>
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
