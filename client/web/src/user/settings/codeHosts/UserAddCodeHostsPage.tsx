import * as H from 'history'
import React from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { CodeHostItem } from './CodeHostItem'
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
    // const [hasDismissedPrivacyWarning, setHasDismissedPrivacyWarning] = useLocalStorage(
    //     'hasDismissedCodeHostPrivacyWarning',
    //     false
    // )
    // const dismissPrivacyWarning = (): void => {
    //     setHasDismissedPrivacyWarning(true)
    // }

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
            {/* {!hasDismissedPrivacyWarning && (
                <div className="alert alert-info">
                    {!userID && (
                        <p>
                            This Sourcegraph installation will never send your code, repository names, file names, or
                            any other specific code data to Sourcegraph.com or any other destination. Your code is kept
                            private on this installation.
                        </p>
                    )}
                    <h3>This Sourcegraph installation will access your code host by:</h3>
                    <ul>
                        <li>
                            Periodically fetching a list of repositories to ensure new, removed, and renamed
                            repositories are accessible on Sourcegraph.
                        </li>
                        <li>Cloning the repositories you specify to create a local cache.</li>
                        <li>Periodically pulling cloned repositories to ensure search results are current.</li>
                        <li>
                            Fetching{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/repo/permissions"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                user repository access permissions
                            </a>
                            , if you have enabled this feature.
                        </li>
                        <li>
                            Opening pull requests and syncing their metadata as part of{' '}
                            <a
                                href="https://docs.sourcegraph.com/user/campaigns"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                code change campaigns
                            </a>
                            , if you have enabled this feature.
                        </li>
                    </ul>
                    <div className="d-flex justify-content-end">
                        <button className="btn btn-light" onClick={dismissPrivacyWarning} type="button">
                            Do not show this again
                        </button>
                    </div>
                </div>
            )} */}

            {codeHostExternalServices && (
                <ul className="list-group">
                    {Object.entries(codeHostExternalServices).map(([id, externalService]) => (
                        <li key={id} className="list-group-item">
                            <CodeHostItem {...externalService} />
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
                            <CodeHostItem {...externalService} />
                        </div>
                    ))}
                </>
            )}
        </div>
    )
}
