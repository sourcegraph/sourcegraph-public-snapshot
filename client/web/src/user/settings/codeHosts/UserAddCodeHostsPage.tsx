import React, { useCallback, useState, useEffect } from 'react'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { CodeHostItem } from './CodeHostItem'
import { PageTitle } from '../../../components/PageTitle'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { ErrorAlert } from '../../../components/alerts'
import { Link } from '../../../../../shared/src/components/Link'
import { isDefined, keyExistsIn } from '../../../../../shared/src/util/types'

import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { SourcegraphContext } from '../../../jscontext'

// type AuthProvider = SourcegraphContext['authProviders'][0]
// type ServiceType = AuthProvider['serviceType']
// type AuthProvidersByType = Partial<Record<ServiceType, AuthProvider>>

export interface UserAddCodeHostsPageProps {
    userID: Scalars['ID']
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    routingPrefix: string
    context: Pick<SourcegraphContext, 'authProviders'>
}

type ServicesByKind = Partial<Record<ExternalServiceKind, ListExternalServiceFields>>
type Status = undefined | 'loading' | ServicesByKind | ErrorLike

const isServicesByKind = (status: Status): status is ServicesByKind =>
    typeof status === 'object' && Object.keys(status).every(key => keyExistsIn(key, ExternalServiceKind))

export const UserAddCodeHostsPage: React.FunctionComponent<UserAddCodeHostsPageProps> = ({
    userID,
    codeHostExternalServices,
    routingPrefix,
    /* context */
}) => {
    const [statusOrError, setStatusOrError] = useState<Status>()
    const [showAddReposFor, setShowAddReposFor] = useState('')

    const fetchExternalServices = useCallback(async () => {
        setStatusOrError('loading')

        const { nodes: fetchedServices } = await queryExternalServices({
            namespace: userID,
            first: null,
            after: null,
        }).toPromise()

        const services: ServicesByKind = fetchedServices.reduce<ServicesByKind>((accumulator, service) => {
            // backend constraint - non-admin users have only one external service per ExternalServiceKind
            accumulator[service.kind] = service
            return accumulator
        }, {})

        setShowAddReposFor('')
        setStatusOrError(services)
    }, [userID])

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsCodeHostConnections')
    }, [])

    useEffect(() => {
        fetchExternalServices().catch(error => {
            setStatusOrError(asError(error))
        })
    }, [fetchExternalServices])

    useEffect(() => {}, [statusOrError])

    const addNewService = useCallback(
        (service: ListExternalServiceFields): void => {
            if (isServicesByKind(statusOrError)) {
                setStatusOrError({ ...statusOrError, [service.kind]: service })
                // display "Add repos" banner only if we just added a connection
                // and there were no errors or warnings
                if (!service.lastSyncError && !service.warning) {
                    setShowAddReposFor(service.displayName)
                }
            }
        },
        [statusOrError]
    )

    const handleError = useCallback((error: ErrorLike): void => {
        // reset 'add your repositories banner', we only want one banner at the
        // time and errors will have it's own
        setShowAddReposFor('')
        setStatusOrError(error)
    }, [])

    const getServiceWarningFragment = ({ id, displayName }: ListExternalServiceFields): React.ReactFragment => (
        <div className="alert alert-danger my-4" key={id}>
            <strong className="align-middle">Could not connect to {displayName}.</strong>
            <span className="align-middle">
                {' '}
                Please remove {displayName} code host connection and try another token to restore the connection.
            </span>
        </div>
    )

    const codeHostOAuthButtons = isServicesByKind(statusOrError)
        ? Object.values(codeHostExternalServices).reduce(
              (accumulator: JSX.Element[], { kind, defaultDisplayName, icon: Icon }) => {
                  if (!statusOrError[kind]) {
                      accumulator.push(
                          <button
                              key={kind}
                              type="button"
                              className={`btn mr-2 ${kind === 'GITLAB' ? 'btn-gitlab' : 'btn-dark'}`}
                          >
                              <Icon className="icon-inline" /> {defaultDisplayName}
                          </button>
                      )
                  }

                  return accumulator
              },
              []
          )
        : []

    return (
        <div className="add-user-code-hosts-page">
            <PageTitle title="Code host connections" />
            <div className="mb-4">
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <h2 className="mb-0">Code host connections</h2>
                </div>
                <p className="text-muted">
                    Connect with your code hosts. Then,{' '}
                    <Link className="text-primary" to={`${routingPrefix}/repositories`}>
                        add repositories
                    </Link>{' '}
                    to search with Sourcegraph.
                </p>
            </div>

            {showAddReposFor && (
                <div className="alert alert-success mb-4" role="alert">
                    Connected with {showAddReposFor}. Next,{' '}
                    <Link className="text-primary" to={`${routingPrefix}/repositories`}>
                        <b>add your repositories →</b>
                    </Link>
                </div>
            )}

            {/* display external service errors */}
            {isServicesByKind(statusOrError) &&
                Object.values(statusOrError)
                    .filter(isDefined)
                    // Services may return warnings/errors immediately or after
                    // the sync. We want to display an alert for both.
                    .filter(service => service.warning || service.lastSyncError)
                    .map(getServiceWarningFragment)}

            {/* display other errors */}
            {isErrorLike(statusOrError) && (
                <ErrorAlert error={statusOrError} prefix="Code host action error" icon={false} />
            )}

            {codeHostExternalServices && isServicesByKind(statusOrError) ? (
                <>
                    {codeHostOAuthButtons.length > 0 && (
                        <div className="border rounded p-4 mb-4">
                            <b>Connect with code host</b>
                            <div className="container">
                                <div className="row py-3">{codeHostOAuthButtons}</div>
                                <div className="row">
                                    <span className="text-muted">
                                        Learn more about{' '}
                                        <Link className="text-primary" to="/will-be-added-soon">
                                            code host connections
                                        </Link>
                                    </span>
                                </div>
                            </div>
                        </div>
                    )}

                    <ul className="list-group">
                        {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) => (
                            <li key={id} className="list-group-item">
                                <CodeHostItem
                                    service={isServicesByKind(statusOrError) ? statusOrError[kind] : undefined}
                                    userID={userID}
                                    kind={kind}
                                    name={defaultDisplayName}
                                    icon={icon}
                                    onDidAdd={addNewService}
                                    onDidRemove={fetchExternalServices}
                                    onDidError={handleError}
                                />
                            </li>
                        ))}
                    </ul>
                </>
            ) : (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner className="icon-inline" />
                </div>
            )}
        </div>
    )
}
