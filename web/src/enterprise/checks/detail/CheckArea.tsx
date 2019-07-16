import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Redirect, Route, Switch } from 'react-router'
import * as sourcegraph from 'sourcegraph'
import { CheckID } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { ThemeProps } from '../../../theme'
import { ThreadActionsArea } from '../../threads/detail/actions/ThreadActionsArea'
import { ChecksAreaContext } from '../scope/ScopeChecksArea'
import { useCheckByTypeForScope } from '../util/useCheckByTypeForScope'
import { CheckDiagnosticsPage } from './diagnostics/CheckDiagnosticsPage'
import { CheckAreaNavbar } from './navbar/CheckAreaNavbar'
import { CheckBreadcrumbs } from './overview/CheckBreadcrumbs'
import { CheckOverview } from './overview/CheckOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends Pick<CheckAreaContext, Exclude<keyof CheckAreaContext, 'checkProvider' | 'checkInfo'>> {}

export interface CheckAreaContext
    extends ChecksAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The check ID. */
    checkID: CheckID

    /**
     * The check provider, without the CheckInformation (which should be accessed on
     * {@link CheckAreaContext#checkInfo}).
     */
    checkProvider: Pick<sourcegraph.CheckProvider, Exclude<keyof sourcegraph.CheckProvider, 'information'>>

    /** The check's information. */
    checkInfo: sourcegraph.CheckInformation

    /** The URL to the check area for this check. */
    checkURL: string

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single check.
 */
export const CheckArea: React.FunctionComponent<Props> = ({ checkID, scope, checkURL, checksURL, ...props }) => {
    const checkOrError = useCheckByTypeForScope(props.extensionsController, checkID, scope)
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
        scope,
        checkID: checkOrError.id,
        checkProvider: checkOrError.provider,
        checkInfo: checkOrError.information,
        checksURL,
        checkURL,
    }

    return (
        <div className="status-area flex-1 d-flex overflow-hidden">
            <div className="flex-1 d-flex flex-column overflow-auto">
                <ErrorBoundary location={props.location}>
                    <div className="container flex-0">
                        <CheckBreadcrumbs
                            checkID={checkID}
                            checkInfo={checkOrError.information}
                            checkURL={checkURL}
                            checksURL={checksURL}
                            className="py-3"
                        />
                        <hr className="my-0" />
                        <h2 className="my-3 font-weight-normal">{checkID.type}</h2>
                    </div>
                    <CheckOverview {...context} className="container flex-0 pb-3" />
                    <div className="w-100 border-bottom" />
                    <CheckAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route path={checkURL} exact={true}>
                            <Redirect to={`${checkURL}/diagnostics`} />
                        </Route>
                        <Route path={`${checkURL}/diagnostics`}>
                            <CheckDiagnosticsPage {...context} />
                        </Route>
                        <Route
                            path={`${checkURL}/settings2`}
                            // tslint:disable-next-line: jsx-no-lambda
                            render={routeComponentProps => (
                                <ThreadActionsArea
                                    {...context}
                                    {...routeComponentProps}
                                    thread={{ url: 'x' }}
                                    threadSettings={{}}
                                    className="container mt-4"
                                />
                            )}
                        />
                        <Route path={`${checkURL}/settings`}>TODO!(sqs)</Route>
                        <Route component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
