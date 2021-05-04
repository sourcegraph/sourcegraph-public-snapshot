import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageHeader } from '../components/PageHeader'
import { VersionContext } from '../schema/site.schema'
import { SearchContextProps } from '../search'

import { SearchContextsListTab } from './SearchContextsListTab'

export interface SearchContextsListPageProps
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'fetchAutoDefinedSearchContexts'> {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    availableVersionContexts: VersionContext[] | undefined
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

export const SearchContextsListPage: React.FunctionComponent<SearchContextsListPageProps> = props => {
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
        <div className="w-100">
            <Page>
                <div className="search-contexts-list-page">
                    <PageHeader
                        path={[
                            {
                                text: 'Search contexts',
                            },
                        ]}
                        className="mb-2"
                    />
                    <p className="text-muted">
                        Search code you care about with search contexts.{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts-experimental"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Learn more
                        </a>
                    </p>
                    <div className="border-bottom my-4">
                        <div className="nav nav-tabs border-bottom-0">
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <a
                                    href=""
                                    role="button"
                                    onClick={onSelectSearchContextsList}
                                    className={classNames('nav-link', selectedTab === 'list' && 'active')}
                                >
                                    Your search contexts
                                </a>
                            </div>
                            {props.authenticatedUser?.siteAdmin && (
                                <div className="nav-item d-flex align-items-center ml-auto">
                                    <Link to="/contexts/convert-version-contexts">Convert version contexts</Link>
                                </div>
                            )}
                        </div>
                    </div>
                    {selectedTab === 'list' && <SearchContextsListTab {...props} />}
                </div>
            </Page>
        </div>
    )
}
