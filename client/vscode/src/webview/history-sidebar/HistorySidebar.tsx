import classNames from 'classnames'
import React, { useEffect, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { CurrentAuthStateResult, CurrentAuthStateVariables } from '../../graphql-operations'
import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { currentAuthStateQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'
import { RecentFile } from './RecentFile'
import { RecentRepo } from './RecentRepo'
import { RecentSearch } from './RecentSearch'

interface HistorySidebarProps extends WebviewPageProps {}

export const HistorySidebar: React.FC<HistorySidebarProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    theme,
}) => {
    const [localRecentSearches, setLocalRecentSearches] = useState<LocalRecentSeachProps[] | undefined>(undefined)
    const [localFileHistory, setLocalFileHistory] = useState<string[] | undefined>(undefined)
    const [authenticatedUser, setAuthenticatedUser] = useState<AuthenticatedUser | null | undefined>(undefined)

    useEffect(() => {
        // Get initial settings
        if (localRecentSearches === undefined) {
            // Get Local Search History
            sourcegraphVSCodeExtensionAPI
                .getLocalRecentSearch()
                .then(response => {
                    setLocalRecentSearches(response)
                })
                .catch(() => {
                    // TODO error handling
                })
            sourcegraphVSCodeExtensionAPI
                .getLocalStorageItem('sg-files-test2')
                .then(response => {
                    setLocalFileHistory(response)
                })
                .catch(() => {
                    // TODO error handling
                })
        }
        if (authenticatedUser === undefined) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                // If user is detected, set valid access token to true
                if (currentUser.data) {
                    setAuthenticatedUser(currentUser.data.currentUser)
                } else {
                    setAuthenticatedUser(null)
                }
            })().catch(() => setAuthenticatedUser(null))
        }
    }, [sourcegraphVSCodeExtensionAPI, localRecentSearches, authenticatedUser, platformContext])

    if (localRecentSearches && authenticatedUser !== undefined) {
        return (
            <div className={styles.sidebarContainer}>
                <RecentSearch
                    localRecentSearches={localRecentSearches}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
                <RecentRepo
                    localRecentSearches={localRecentSearches}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
                <RecentFile
                    localFileHistory={localFileHistory}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
                {!authenticatedUser && (
                    <div className={styles.sidebarSection}>
                        <h5 className="flex-grow-1 btn-outline-secondary my-2">Search Your Private Code</h5>
                        <div className={classNames('p-1', styles.sidebarSectionList)}>
                            <div className={classNames('p-1', styles.sidebarSectionListItem)}>
                                <p className={classNames('mt-1 mb-3 text')}>
                                    Create an account to enhance search across your private repositories: search
                                    multiple repos & commit history, monitor, save searches, and more.
                                </p>
                            </div>
                            <div className={classNames('p-1', styles.sidebarSectionListItem)}>
                                <button
                                    type="submit"
                                    className={classNames(
                                        'btn btn-sm btn-primary btn-link w-100 border-0 font-weight-normal'
                                    )}
                                >
                                    <span className="py-1">Create an account</span>
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        )
    }
    return <LoadingSpinner />
}
