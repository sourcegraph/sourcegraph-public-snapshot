import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckAreaContext } from '../CheckArea'
import { CheckDiagnosticGroup } from './CheckDiagnosticGroup'
import { useCheckDiagnosticGroups } from './useCheckDiagnosticGroups'

interface Props extends Pick<CheckAreaContext, 'checkID' | 'checkProvider' | 'checkInfo'>, ExtensionsControllerProps {
    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The check diagnostics page.
 */
export const CheckDiagnosticsPage: React.FunctionComponent<Props> = ({
    checkID,
    checkProvider,
    checkInfo,
    className = '',
    itemClassName = '',
    ...props
}) => {
    const diagnosticGroupsOrError = useCheckDiagnosticGroups(props.extensionsController, checkProvider)
    return (
        <div className={`check-diagnostics-page ${className}`}>
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
                    <p className="p-2 mb-0 text-muted">No notifications found.</p>
                </div>
            ) : (
                <ul className="list-unstyled mb-0">
                    {diagnosticGroupsOrError.map((notification, i) => (
                        <li key={i}>
                            <CheckDiagnosticGroup
                                {...props}
                                diagnosticGroup={notification}
                                className="card my-5"
                                contentClassName="card-body"
                            />
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}
