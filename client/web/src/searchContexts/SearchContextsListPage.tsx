import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { VersionContext } from '../schema/site.schema'
import { isSearchContextSpecAvailable } from '../search'
import { convertVersionContextToSearchContext } from '../search/backend'

import { ConvertVersionContextsTab } from './ConvertVersionContextsTab'
import { SearchContextsListTab } from './SearchContextsListTab'

export interface SearchContextsListPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    availableVersionContexts: VersionContext[] | undefined
}

type SelectedTab = 'list' | 'convert-version-contexts'

function getSelectedTabFromLocation(locationSearch: string): SelectedTab {
    const urlParameters = new URLSearchParams(locationSearch)
    switch (urlParameters.get('tab')) {
        case 'list':
            return 'list'
        case 'convert-version-contexts':
            return 'convert-version-contexts'
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

    const onSelectConvertVersionContexts = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setTab('convert-version-contexts')
        },
        [setTab]
    )

    return (
        <div className="w-100">
            <Page>
                <div className="search-contexts-list-page">
                    <div className="search-contexts-list-page__title mb-3">
                        <h2>Search contexts</h2>
                    </div>
                    <div className="border-bottom mb-4">
                        <div className="nav nav-tabs border-bottom-0">
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <a
                                    href=""
                                    role="button"
                                    onClick={onSelectSearchContextsList}
                                    className={classNames('nav-link', selectedTab === 'list' && 'active')}
                                >
                                    Search contexts
                                </a>
                            </div>
                            {props.authenticatedUser?.siteAdmin && (
                                <div className="nav-item test-convert-version-contexts-tab">
                                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                    <a
                                        href=""
                                        role="button"
                                        onClick={onSelectConvertVersionContexts}
                                        className={classNames(
                                            'nav-link',
                                            selectedTab === 'convert-version-contexts' && 'active'
                                        )}
                                    >
                                        Convert version contexts
                                    </a>
                                </div>
                            )}
                        </div>
                    </div>
                    {selectedTab === 'list' && <SearchContextsListTab {...props} />}
                    {selectedTab === 'convert-version-contexts' && (
                        <ConvertVersionContextsTab
                            {...props}
                            isSearchContextSpecAvailable={isSearchContextSpecAvailable}
                            convertVersionContextToSearchContext={convertVersionContextToSearchContext}
                        />
                    )}
                </div>
            </Page>
        </div>
    )
}
