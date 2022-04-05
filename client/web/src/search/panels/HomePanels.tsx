import * as React from 'react'
import { useEffect, useCallback } from 'react'

import { ApolloQueryResult, gql } from '@apollo/client'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Tabs, Tab, TabList, TabPanel, TabPanels } from '@sourcegraph/wildcard'

import { HomePanelsProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { CollaboratorsFragment, HomePanelsQueryResult, HomePanelsQueryVariables } from '../../graphql-operations'
import { GettingStartedTour } from '../../tour/GettingStartedTour'
import { eventLogger } from '../../tracking/eventLogger'

import { CollaboratorsPanel } from './CollaboratorsPanel'
import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'
import {
    collaboratorsFragment,
    recentlySearchedRepositoriesFragment,
    recentFilesFragment,
    recentSearchesPanelFragment,
    savedSearchesPanelFragment,
} from './PanelFragments'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

import styles from './HomePanels.module.scss'

interface Props extends TelemetryProps, HomePanelsProps, FeatureFlagProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    showCollaborators: boolean
}

const INVITES_TAB_KEY = 'homepage.userInvites.tab'

// Use a larger page size because not every search may have a `repo:` filter, and `repo:` filters could often
// be duplicated. Therefore, we fetch more searches to populate this panel.
export const RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD = 50
export const RECENT_SEARCHES_TO_LOAD = 20
export const RECENT_FILES_TO_LOAD = 20

export type HomePanelsFetchMore = (
    fetchMoreOptions: Partial<HomePanelsQueryVariables>
) => Promise<ApolloQueryResult<HomePanelsQueryResult>>

export const HOME_PANELS_QUERY = gql`
    query HomePanelsQuery(
        $userId: ID!
        $firstRecentlySearchedRepositories: Int!
        $firstRecentSearches: Int!
        $firstRecentFiles: Int!
        $enableSavedSearches: Boolean!
        $enableCollaborators: Boolean!
    ) {
        node(id: $userId) {
            __typename
            ...RecentlySearchedRepositoriesFragment
            ...RecentSearchesPanelFragment
            ...RecentFilesFragment
            ...CollaboratorsFragment
        }
        ...SavedSearchesPanelFragment
    }
    ${recentlySearchedRepositoriesFragment}
    ${recentSearchesPanelFragment}
    ${savedSearchesPanelFragment}
    ${recentFilesFragment}
    ${collaboratorsFragment}
`

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => {
    const userId = props.authenticatedUser?.id || ''
    const showCollaborators = props.showCollaborators
    const showSavedSearches = !props.isSourcegraphDotCom

    const { data, fetchMore: rawFetchMore } = useQuery<HomePanelsQueryResult, HomePanelsQueryVariables>(
        HOME_PANELS_QUERY,
        {
            variables: {
                userId,
                firstRecentlySearchedRepositories: RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD,
                firstRecentSearches: RECENT_SEARCHES_TO_LOAD,
                firstRecentFiles: RECENT_FILES_TO_LOAD,
                enableSavedSearches: showSavedSearches,
                enableCollaborators: showCollaborators,
            },
        }
    )

    const fetchMore: HomePanelsFetchMore = useCallback(
        (variables: Partial<HomePanelsQueryVariables>) =>
            rawFetchMore({
                variables: {
                    userId,
                    firstRecentlySearchedRepositories: 0,
                    firstRecentSearches: 0,
                    firstRecentFiles: 0,
                    enableSavedSearches: false,
                    enableCollaborators: false,
                    ...variables,
                },
            }),
        [rawFetchMore, userId]
    )

    useEffect(() => {
        if (props.showCollaborators === true) {
            return
        }
        const loggerPayload = {
            // The other types are emitted in <CollaboratorsPanel />
            type: 'config-disabled',
        }
        eventLogger.log('HomepageInvitationsViewEmpty', loggerPayload, loggerPayload)
    }, [props.showCollaborators])

    const node = data?.node ?? null
    if (node !== null && node.__typename !== 'User') {
        return null
    }

    return (
        <div className={classNames('container', styles.homePanels)} data-testid="home-panels">
            <div className="row">
                <GettingStartedTour
                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                    telemetryService={props.telemetryService}
                    featureFlags={props.featureFlags}
                    isAuthenticated={true}
                    className="w-100"
                    variant="horizontal"
                />
            </div>
            <div className="row">
                <RepositoriesPanel
                    {...props}
                    className={classNames('col-lg-4', styles.panel)}
                    recentlySearchedRepositories={node}
                    fetchMore={fetchMore}
                />
                <RecentSearchesPanel
                    {...props}
                    recentSearches={node}
                    className={classNames('col-lg-8', styles.panel)}
                    fetchMore={fetchMore}
                />
            </div>
            <div className="row">
                <RecentFilesPanel
                    {...props}
                    className={classNames('col-lg-7', styles.panel)}
                    recentFilesFragment={node}
                    fetchMore={fetchMore}
                />

                {showCollaborators ? (
                    <CollaboratorsTabPanel {...props} data={data} collaboratorsFragment={node} />
                ) : showSavedSearches ? (
                    <SavedSearchesPanel
                        {...props}
                        className={classNames('col-lg-5', styles.panel)}
                        savedSearchesFragment={data ?? null}
                    />
                ) : (
                    <CommunitySearchContextsPanel {...props} className={classNames('col-lg-5', styles.panel)} />
                )}
            </div>
        </div>
    )
}

interface CollaboratorsTabPanelProps extends Props {
    data: undefined | HomePanelsQueryResult
    collaboratorsFragment: null | CollaboratorsFragment
}

const CollaboratorsTabPanel: React.FunctionComponent<CollaboratorsTabPanelProps> = ({
    data,
    collaboratorsFragment,
    ...props
}) => {
    const [persistedTabIndex, setPersistedTabIndex] = useTemporarySetting(INVITES_TAB_KEY, 1)

    if (persistedTabIndex === undefined) {
        return null
    }

    return (
        <div className={classNames('col-lg-5', styles.panel)}>
            <Tabs defaultIndex={persistedTabIndex} onChange={setPersistedTabIndex} className={styles.tabs}>
                <TabList>
                    <Tab>{props.isSourcegraphDotCom ? 'Community search contexts' : 'Saved searches'}</Tab>
                    <Tab>Invite colleagues</Tab>
                </TabList>
                <TabPanels className={classNames('h-100', styles.tabPanels)}>
                    <TabPanel className="h-100">
                        {props.isSourcegraphDotCom ? (
                            <CommunitySearchContextsPanel {...props} insideTabPanel={true} />
                        ) : (
                            <SavedSearchesPanel {...props} insideTabPanel={true} savedSearchesFragment={data ?? null} />
                        )}
                    </TabPanel>
                    <TabPanel className="h-100">
                        <CollaboratorsPanel {...props} collaboratorsFragment={collaboratorsFragment} />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>
    )
}
