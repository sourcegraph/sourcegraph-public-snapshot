import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../../components/ErrorBoundary'
import { ThemeProps } from '../../../../theme'
import { CheckAreaContext } from '../CheckArea'
import { CheckDiagnosticGroupPage } from './detail/CheckDiagnosticGroupPage'
import { CheckDiagnosticGroupsList } from './list/CheckDiagnosticGroupsList'
import { useCheckDiagnosticGroups } from './list/useCheckDiagnosticGroups'

interface Props
    extends Pick<CheckAreaContext, 'checkID' | 'checkProvider' | 'checkInfo'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    checkDiagnosticsURL: string

    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The check diagnostics area.
 */
export const CheckDiagnosticsArea: React.FunctionComponent<Props> = ({
    checkID,
    checkProvider,
    checkInfo,
    checkDiagnosticsURL,
    className = '',
    itemClassName = '',
    ...props
}) => {
    const diagnosticGroupsOrError = useCheckDiagnosticGroups(props.extensionsController, checkProvider)
    return (
        <div className={`check-diagnostics-area d-flex align-items-start ${className}`}>
            {isErrorLike(diagnosticGroupsOrError) ? (
                <div className={itemClassName}>
                    <div className="alert alert-danger mt-2">{diagnosticGroupsOrError.message}</div>
                </div>
            ) : diagnosticGroupsOrError === LOADING ? (
                <div className={itemClassName}>
                    <LoadingSpinner className="mt-3" />
                </div>
            ) : diagnosticGroupsOrError.length === 0 ? (
                <div className={itemClassName}>
                    <p className="p-2 mb-0 text-muted">No diagnostics found.</p>
                </div>
            ) : (
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route  path={checkDiagnosticsURL} exact={true}>
                            <CheckDiagnosticGroupsList
                                {...props}
                                checkProvider={checkProvider}
                                diagnosticGroups={diagnosticGroupsOrError}
                                checkDiagnosticsURL={checkDiagnosticsURL}
                                className="flex-1"
                                itemClassName="my-5"
                            />
                        </Route>
                        <Route
                            
                            path={`${checkDiagnosticsURL}/:id`}
                            exact={true}
                            // tslint:disable-next-line: jsx-no-lambda
                            render={(routeComponentProps: RouteComponentProps<{ id: string }>) => {
                                const diagnosticGroup = diagnosticGroupsOrError.find(
                                    ({ id }) => routeComponentProps.match.params.id === id
                                )
                                return diagnosticGroup ? (
                                    <CheckDiagnosticGroupPage
                                        {...props}
                                        checkProvider={checkProvider}
                                        diagnosticGroup={diagnosticGroup}
                                        checkDiagnosticsURL={checkDiagnosticsURL}
                                    />
                                ) : (
                                    <div className="alert alert-danger mt-2">
                                        The requested diagnostic group was not found.
                                    </div>
                                )
                            }}
                        />
                    </Switch>
                </ErrorBoundary>
            )}
        </div>
    )
}
