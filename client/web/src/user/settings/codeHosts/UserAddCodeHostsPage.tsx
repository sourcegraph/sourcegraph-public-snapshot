import React from 'react'
import * as H from 'history'

import { CodeHostItem } from './CodeHostItem'
import { PageTitle } from '../../../components/PageTitle'
import { AddExternalServiceOptions } from '../../../components/externalServices/externalServices'
import { Link } from '../../../../../shared/src/components/Link'
// import { useLocalStorage } from '../../../util/useLocalStorage'
import { Scalars } from '../../../graphql-operations'

export interface UserAddCodeHostsPageProps {
    history: H.History
    userID?: Scalars['ID']

    /**
     * The list of code host external services to be displayed.
     * Pick items from externalServices.codeHostExternalServices.
     */
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    /**
     * The list of non-code host external services to be displayed.
     * Pick items from externalServices.nonCodeHostExternalServices.
     */
    nonCodeHostExternalServices: Record<string, AddExternalServiceOptions>
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const UserAddCodeHostsPage: React.FunctionComponent<UserAddCodeHostsPageProps> = ({
    codeHostExternalServices,
    nonCodeHostExternalServices,
    /* userID, */
}) => {
    console.log('here')

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
            {true && (
                <div className="alert alert-warning my-4">
                    <strong>Could not connect to GitHub.</strong> Please <Link to="/">update your access token</Link> to
                    restore the connection.
                </div>
            )}
            {codeHostExternalServices && (
                <ul className="list-group">
                    {Object.entries(codeHostExternalServices).map(([id, { kind, defaultDisplayName, icon }]) => (
                        <li key={id} className="list-group-item">
                            <CodeHostItem
                                icon={icon}
                                name={defaultDisplayName}
                                kind={kind}
                                onDidConnect={() => {}}
                                onDidRemove={() => {}}
                                onDidEdit={() => {}}
                            />
                        </li>
                    ))}
                </ul>
            )}

            {Object.entries(nonCodeHostExternalServices).length > 0 && (
                <>
                    <br />
                    <h2>Other connections</h2>
                    <p className="mt-2">Add connections to non-code-host services.</p>
                    {Object.entries(nonCodeHostExternalServices).map(([id, externalService]) => (
                        <div className="add-external-services-page__card" key={id}>
                            <CodeHostItem
                                onDidConnect={() => {}}
                                onDidRemove={() => {}}
                                onDidEdit={() => {}}
                                {...externalService}
                            />
                        </div>
                    ))}
                </>
            )}
        </div>
    )
}
