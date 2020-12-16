import React, { useCallback, useState, useEffect } from 'react'
import * as H from 'history'

import { CodeHostItem } from './CodeHostItem'
import { UpdateCodeHostConnectionModal } from './UpdateCodeHostConnectionModal'
import { hints } from './modalHints'
import { PageTitle } from '../../../components/PageTitle'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { Link } from '../../../../../shared/src/components/Link'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { eventLogger } from '../../../tracking/eventLogger'

import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { queryExternalServices } from '../../../components/externalServices/backend'
import { ErrorAlert } from '../../../components/alerts'

export interface UserAddCodeHostsPageProps {
    userID: Scalars['ID']
    codeHostExternalServices: Record<ExternalServiceKind, AddExternalServiceOptions>
    history: H.History
}

type Status = undefined | 'loading' | ErrorLike
export type servicesByKindState = Partial<Record<ExternalServiceKind, ListExternalServiceFields>>

export const UserAddCodeHostsPage: React.FunctionComponent<UserAddCodeHostsPageProps> = ({
    userID,
    codeHostExternalServices,
    history,
}) => {
    const [servicesByKind, setServicesByKind] = useState<servicesByKindState>({})
    const [statusOrError, setStatusOrError] = useState<Status>()
    const [showUpdateConnectionModal, setShowUpdateConnectionModal] = useState(false)
    const toggleUpdateConnectionModal = useCallback(() => setShowUpdateConnectionModal(!showUpdateConnectionModal), [
        showUpdateConnectionModal,
    ])

    const fetchExternalServices = useCallback(async () => {
        setStatusOrError('loading')

        const { nodes: fetchedServices } = await queryExternalServices({
            namespace: userID,
            first: null,
            after: null,
        }).toPromise()

        const services: servicesByKindState = fetchedServices.reduce<servicesByKindState>((accumulator, service) => {
            accumulator[service.kind] = service
            return accumulator
        }, {})

        setServicesByKind(services)
        setStatusOrError(undefined)
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
            setServicesByKind({ ...servicesByKind, [service.kind]: service })
        },
        [servicesByKind]
    )

    const getServiceWarningFragment = ({
        displayName,
        kind,
        id,
        warning,
        config,
    }: ListExternalServiceFields): React.ReactFragment => (
        <div className="alert alert-danger my-4" key={id}>
            <strong className="align-middle">Could not connect to {displayName}.</strong>
            <span className="align-middle"> Please </span>
            <button
                type="button"
                className="btn btn-link text-primary p-0 shadow-none"
                onClick={toggleUpdateConnectionModal}
            >
                update your access token
            </button>{' '}
            <span className="align-middle">to restore the connection.</span>
            <div className="py-2">
                <small>{warning}</small>
            </div>
            {showUpdateConnectionModal && (
                <UpdateCodeHostConnectionModal
                    serviceId={id}
                    serviceConfig={config}
                    name={displayName}
                    kind={kind}
                    hintFragment={hints[kind]}
                    onDidCancel={toggleUpdateConnectionModal}
                    onDidUpdate={handleServiceUpsert}
                    onDidError={setStatusOrError}
                />
            )}
        </div>
    )

    return (
        <div className="add-user-code-hosts-page">
            <PageTitle title="Code host connections" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2 className="mb-0">Code host connections</h2>
            </div>
            <p className="text-muted mt-2">
                Connect with providers where your source code is hosted. Then,{' '}
                <Link className="text-primary" to="repositories">
                    add repositories
                </Link>{' '}
                to search with Sourcegraph.
            </p>

            {Object.values(servicesByKind)
                .filter(service => service?.warning)
                .map(getServiceWarningFragment)}

            {isErrorLike(statusOrError) && (
                <ErrorAlert error={statusOrError} history={history} prefix="Code host action error" icon={false} />
            )}

            {codeHostExternalServices && statusOrError !== 'loading' && (
                <ul className="list-group">
                    {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) => (
                        <li key={id} className="list-group-item">
                            <CodeHostItem
                                service={servicesByKind[kind]}
                                userID={userID}
                                kind={kind}
                                name={defaultDisplayName}
                                icon={icon}
                                onDidConnect={handleServiceUpsert}
                                onDidRemove={fetchExternalServices}
                                onDidError={setStatusOrError}
                            />
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}
