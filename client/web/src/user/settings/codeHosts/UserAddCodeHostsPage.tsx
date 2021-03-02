import React, { useCallback, useState, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import * as H from 'history'

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

type AuthProvider = SourcegraphContext['authProviders'][0]
type ServiceType = AuthProvider['serviceType']
type AuthProvidersByType = Partial<Record<ServiceType, AuthProvider>>

export interface UserAddCodeHostsPageProps extends RouteComponentProps {
    userID: Scalars['ID']
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    history: H.History
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
    history,
    routingPrefix,
    context,
}) => {
    const [statusOrError, setStatusOrError] = useState<Status>()

    const [isUpdateModalOpen, setIssUpdateModalOpen] = useState(false)
    const toggleUpdateModal = useCallback(() => {
        setIssUpdateModalOpen(!isUpdateModalOpen)
    }, [isUpdateModalOpen])

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

    const handleServiceUpsert = useCallback(
        (service: ListExternalServiceFields): void => {
            if (isServicesByKind(statusOrError)) {
                setStatusOrError({ ...statusOrError, [service.kind]: service })
            }
        },
        [statusOrError]
    )

    const getServiceWarningFragment = ({ id, displayName }: ListExternalServiceFields): React.ReactFragment => (
        <div className="alert alert-danger my-4" key={id}>
            <strong className="align-middle">Could not connect to {displayName}.</strong>
            <span className="align-middle"> Please </span>
            <button type="button" className="btn btn-link text-primary p-0" onClick={toggleUpdateModal}>
                update your access token
            </button>{' '}
            <span className="align-middle">to restore the connection.</span>
        </div>
    )

    const authProvidersByType = context.authProviders.reduce((accumulator: AuthProvidersByType, provider) => {
        accumulator[provider.serviceType] = provider
        return accumulator
    }, {})

    const sendOAuthRequest = (kind: ExternalServiceKind): void => {
        const lowerCaseKind = kind.toLocaleLowerCase()
        const authProvider = authProvidersByType[lowerCaseKind]
        debugger
        window.location.assign(`${authProvider.authenticationURL as string}&redirect=${window.location.href}`)
    }

    const codeHostOAuthButtons = isServicesByKind(statusOrError)
        ? Object.values(codeHostExternalServices).reduce(
              (accumulator: JSX.Element[], { kind, defaultDisplayName, icon: Icon }) => {
                  if (!statusOrError[kind]) {
                      accumulator.push(
                          <button
                              onClick={() => sendOAuthRequest(kind)}
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
                <ErrorAlert error={statusOrError} history={history} prefix="Code host action error" icon={false} />
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
                                        <Link className="text-primary" to="/tbd">
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
                                    isUpdateModalOpen={isUpdateModalOpen}
                                    toggleUpdateModal={toggleUpdateModal}
                                    onDidUpsert={handleServiceUpsert}
                                    onDidRemove={fetchExternalServices}
                                    onDidError={setStatusOrError}
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
