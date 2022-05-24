import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PlusIcon from 'mdi-react/PlusIcon'

import { SearchContextProps } from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { PageHeader, Link, Button, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'

import { SearchContextsListTab } from './SearchContextsListTab'

export interface SearchContextsListPageProps
    extends Pick<
            SearchContextProps,
            'fetchSearchContexts' | 'fetchAutoDefinedSearchContexts' | 'getUserSearchContextNamespaces'
        >,
        PlatformContextProps<'requestGraphQL'> {
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

type SelectedTab = 'list'

function getSelectedTabFromLocation(locationSearch: string): SelectedTab {
    const urlParameters = new URLSearchParams(locationSearch)
    switch (urlParameters.get('tab')) {
        case 'list':
            return 'list'
    }
    return 'list'
}

function setSelectedLocationTab(location: H.Location, history: H.History, selectedTab: SelectedTab): void {
    const urlParameters = new URLSearchParams(location.search)
    urlParameters.set('tab', selectedTab)
    if (location.search !== urlParameters.toString()) {
        history.replace({ ...location, search: urlParameters.toString() })
    }
}

export const SearchContextsListPage: React.FunctionComponent<
    React.PropsWithChildren<SearchContextsListPageProps>
> = props => {
    const [selectedTab, setSelectedTab] = useState<SelectedTab>(getSelectedTabFromLocation(props.location.search))

    const setTab = useCallback(
        (tab: SelectedTab) => {
            setSelectedTab(tab)
            setSelectedLocationTab(props.location, props.history, tab)
        },
        [props.location, props.history]
    )

    const onSelectSearchContextsList = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setTab('list')
        },
        [setTab]
    )

    return (
        <div data-testid="search-contexts-list-page" className="w-100">
            <Page>
                <PageHeader
                    path={[
                        {
                            icon: MagnifyIcon,
                            to: '/search',
                            ariaLabel: 'Code Search',
                        },
                        {
                            text: 'Contexts',
                        },
                    ]}
                    actions={
                        <Button to="/contexts/new" variant="primary" as={Link}>
                            <Icon as={PlusIcon} />
                            Create search context
                        </Button>
                    }
                    description={
                        <span className="text-muted">
                            Search code you care about with search contexts.{' '}
                            <Link
                                to="/help/code_search/explanations/features#search-contexts"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Learn more
                            </Link>
                        </span>
                    }
                    className="mb-3"
                />
                <div className="mb-4">
                    <div className="nav nav-tabs">
                        <div className="nav-item">
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <Link
                                to=""
                                role="button"
                                onClick={onSelectSearchContextsList}
                                className={classNames('nav-link', selectedTab === 'list' && 'active')}
                            >
                                <span className="text-content" data-tab-content="Your search contexts">
                                    Your search contexts
                                </span>
                            </Link>
                        </div>
                    </div>
                </div>
                {selectedTab === 'list' && <SearchContextsListTab {...props} />}
            </Page>
        </div>
    )
}
