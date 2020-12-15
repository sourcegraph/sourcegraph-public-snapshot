import React, { useCallback, useState, useEffect } from 'react'
import * as H from 'history'

import { CodeHostItem } from './CodeHostItem'
import { PageTitle } from '../../../components/PageTitle'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { Link } from '../../../../../shared/src/components/Link'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { eventLogger } from '../../../tracking/eventLogger'

import { Scalars, ExternalServiceKind, ExternalServiceFields } from '../../../graphql-operations'
import { queryExternalServices } from '../../../components/externalServices/backend'

export interface UserAddCodeHostsPageProps {
    history: H.History
    userID: Scalars['ID']
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
}

type Status = undefined | 'loading' | 'loaded' | ErrorLike
export type servicesByKindState = Partial<
    Record<
        ExternalServiceKind,
        { serviceID: ExternalServiceFields['id']; repoCount?: number; warning?: string } /* []*/
    >
>

export const UserAddCodeHostsPage: React.FunctionComponent<UserAddCodeHostsPageProps> = ({
    userID,
    codeHostExternalServices,
}) => {
    const [servicesByKind, setServicesByKind] = useState<servicesByKindState>({})
    const [statusOrError, setStatusOrError] = useState<Status>()

    const fetchExternalServices = useCallback(async () => {
        setStatusOrError('loading')

        const { nodes: fetchedServices } = await queryExternalServices({
            namespace: userID,
            first: null,
            after: null,
        }).toPromise()

        const services: servicesByKindState = fetchedServices.reduce((accumulator, { id, kind }) => {
            // TODO: Figure out what to do when user has multiple external services of the same kind.
            // Is it possible by design?

            // const byKind = accumulator[kind]
            // if (!byKind) {
            //     accumulator[kind] = [id]
            // } else {
            //     byKind.push(id)
            // }
            accumulator[kind] = { serviceID: id }

            return accumulator
        }, {} as servicesByKindState)

        setServicesByKind(services)
        setStatusOrError('loaded')
    }, [userID])

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsCodeHostConnections')
    }, [])

    useEffect(() => {
        fetchExternalServices().catch(error => {
            setStatusOrError(asError(error))
        })
    }, [fetchExternalServices])

    return (
        <div className="add-user-code-hosts-page">
            <PageTitle title="Code host connections" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2 className="mb-0">Code host connections</h2>
            </div>
            <p className="text-muted mt-2">
                Connect with providers where your source code is hosted. Then,{' '}
                <Link to="repositories">add repositories</Link> to search with Sourcegraph.
            </p>
            {isErrorLike(statusOrError) && (
                <div className="alert alert-danger my-4">
                    <div className="pb-2">{statusOrError.message}</div>
                    <strong>Could not connect to GitHub.</strong> Please <Link to="/">update your access token</Link> to
                    restore the connection.
                </div>
            )}
            {codeHostExternalServices && statusOrError !== 'loading' && (
                <ul className="list-group">
                    {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) => (
                        <li key={id} className="list-group-item">
                            <CodeHostItem
                                {...servicesByKind[kind]}
                                userID={userID}
                                kind={kind}
                                name={defaultDisplayName}
                                icon={icon}
                                onDidConnect={fetchExternalServices}
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
