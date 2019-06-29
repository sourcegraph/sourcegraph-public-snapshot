import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, Switch } from 'react-router'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { Status } from '../status'
import { useStatusByTypeForScope } from '../util/useStatusByTypeForScope'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends Pick<StatusAreaContext, Exclude<keyof StatusAreaContext, 'status'>> {}

export interface StatusAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The status type. */
    type: string

    /** The status scope. */
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot

    /** The status. */
    status: Status

    areaURL: string
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single status.
 */
export const StatusArea: React.FunctionComponent<Props> = ({ type, scope, areaURL, ...props }) => {
    const statusOrError = useStatusByTypeForScope(props.extensionsController, type, scope)
    if (statusOrError === LOADING) {
        return null // loading
    }
    if (statusOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Status not found" />
    }
    if (isErrorLike(statusOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={statusOrError.message} />
    }

    const context: StatusAreaContext & {
        areaURL: string
    } = {
        ...props,
        type,
        scope,
        areaURL,
        status: statusOrError,
    }

    return (
        <div className="status-area">
            <ErrorBoundary location={props.location}>
                <Switch>
                    <Route
                        path={areaURL}
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <p>
                                TODO!(sqs) {context} {routeComponentProps}
                            </p>
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </ErrorBoundary>
        </div>
    )
}
