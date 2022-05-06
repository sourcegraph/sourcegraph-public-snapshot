import React, { useCallback, useEffect, useState } from 'react'

import { gql } from '@apollo/client'
import classNames from 'classnames'
import PencilOutlineIcon from 'mdi-react/PencilOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, ButtonGroup, Link, Menu, MenuButton, MenuList, MenuItem, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { SavedSearchesPanelFragment } from '../../graphql-operations'
import { buildSearchURLQueryFromQueryState } from '../../stores'

import { EmptyPanelContainer } from './EmptyPanelContainer'
import { FooterPanel } from './FooterPanel'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    savedSearchesFragment: SavedSearchesPanelFragment | null
    insideTabPanel?: boolean
}

export const savedSearchesPanelFragment = gql`
    fragment SavedSearchesPanelFragment on Query {
        savedSearches @include(if: $enableSavedSearches) {
            id
            description
            notify
            notifySlack
            query
            namespace {
                __typename
                id
                namespaceName
            }
            slackWebhookURL
        }
    }
`

export const SavedSearchesPanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    className,
    telemetryService,
    insideTabPanel,
    savedSearchesFragment,
}) => {
    const savedSearches = savedSearchesFragment?.savedSearches ?? null

    const [showAllSearches, setShowAllSearches] = useState(true)

    useEffect(() => {
        // Only log the first load (when items to load is equal to the page size)
        if (savedSearches) {
            telemetryService.log(
                'SavedSearchesPanelLoaded',
                { empty: savedSearches.length === 0, showAllSearches },
                { empty: savedSearches.length === 0, showAllSearches }
            )
        }
    }, [savedSearches, telemetryService, showAllSearches])

    const logEvent = useCallback((event: string, props?: any) => (): void => telemetryService.log(event, props), [
        telemetryService,
    ])

    const emptyDisplay = (
        <EmptyPanelContainer className="text-muted">
            <small>
                Use saved searches to alert you to uses of a favorite API, or changes to code you need to monitor.
            </small>
            {authenticatedUser && (
                <Button
                    to={`/users/${authenticatedUser.username}/searches/add`}
                    onClick={logEvent('SavedSearchesPanelCreateButtonClicked', { source: 'empty view' })}
                    className="mt-2 align-self-center"
                    variant="secondary"
                    as={Link}
                >
                    <Icon as={PlusIcon} />
                    Create a saved search
                </Button>
            )}
        </EmptyPanelContainer>
    )
    const loadingDisplay = <LoadingPanelView text="Loading saved searches" />

    const contentDisplay = (
        <div className="d-flex flex-column h-100">
            <div className="d-flex justify-content-between mb-1 mt-2">
                <small>Search</small>
                <small>Edit</small>
            </div>
            <ul className="list-group-flush flex-grow-1 list-group mb-3">
                {savedSearches
                    ?.filter(search => (showAllSearches ? true : search.namespace.id === authenticatedUser?.id))
                    .map(search => (
                        <li key={search.id} className="text-monospace test-saved-search-entry d-block mb-2">
                            <div className="d-flex justify-content-between">
                                <small>
                                    <Link
                                        to={'/search?' + buildSearchURLQueryFromQueryState({ query: search.query })}
                                        className="p-0"
                                        onClick={logEvent('SavedSearchesPanelSearchClicked')}
                                    >
                                        {search.description}
                                    </Link>
                                </small>
                                {authenticatedUser &&
                                    (search.namespace.__typename === 'User' ? (
                                        <Link
                                            to={`/users/${search.namespace.namespaceName}/searches/${search.id}`}
                                            onClick={logEvent('SavedSearchesPanelEditClicked')}
                                            aria-label={`Edit saved search ${search.description}`}
                                        >
                                            <Icon as={PencilOutlineIcon} />
                                        </Link>
                                    ) : (
                                        <Link
                                            to={`/organizations/${search.namespace.namespaceName}/searches/${search.id}`}
                                            onClick={logEvent('SavedSearchesPanelEditClicked')}
                                            aria-label={`Edit saved search ${search.description}`}
                                        >
                                            <Icon as={PencilOutlineIcon} />
                                        </Link>
                                    ))}
                            </div>
                        </li>
                    ))}
            </ul>
            {authenticatedUser && (
                <FooterPanel className="p-1">
                    <small>
                        {/*
                           a11y-ignore
                           Rule: "color-contrast" (Elements must have sufficient color contrast)
                        */}
                        <Link
                            to={`/users/${authenticatedUser.username}/searches`}
                            className="text-left a11y-ignore"
                            onClick={logEvent('SavedSearchesPanelViewAllClicked')}
                        >
                            View saved searches
                        </Link>
                    </small>
                </FooterPanel>
            )}
        </div>
    )

    const actionButtons = (
        <>
            <ButtonGroup className="d-none d-sm-block d-lg-none d-xl-block">
                <Button
                    onClick={() => setShowAllSearches(false)}
                    className="test-saved-search-panel-my-searches"
                    outline={showAllSearches}
                    variant="secondary"
                    size="sm"
                >
                    My searches
                </Button>
                <Button
                    onClick={() => setShowAllSearches(true)}
                    className="test-saved-search-panel-all-searches"
                    outline={!showAllSearches}
                    variant="secondary"
                    size="sm"
                >
                    All searches
                </Button>
            </ButtonGroup>
            <Menu>
                <MenuButton
                    variant="icon"
                    outline={true}
                    className="d-block d-sm-none d-lg-block d-xl-none p-0"
                    size="lg"
                    aria-label="Filter saved searches"
                >
                    ...
                </MenuButton>

                <MenuList>
                    <MenuItem onSelect={() => setShowAllSearches(false)}>My searches</MenuItem>
                    <MenuItem onSelect={() => setShowAllSearches(true)}>All searches</MenuItem>
                </MenuList>
            </Menu>
        </>
    )
    return (
        <PanelContainer
            insideTabPanel={insideTabPanel}
            title="Saved searches"
            className={classNames(className, { 'h-100': insideTabPanel })}
            state={savedSearches ? (savedSearches.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
