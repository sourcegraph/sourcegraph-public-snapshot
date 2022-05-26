import React from 'react'

import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PlusIcon from 'mdi-react/PlusIcon'

import { SearchContextProps } from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { PageHeader, Link, Button, Icon, Tabs, Tab, TabList, TabPanel, TabPanels } from '@sourcegraph/wildcard'

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

export const SearchContextsListPage: React.FunctionComponent<
    React.PropsWithChildren<SearchContextsListPageProps>
> = props => (
    <div data-testid="search-contexts-list-page" className="w-100">
        <Page>
            <PageHeader
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
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={MagnifyIcon} to="/search" aria-label="Code Search" />
                    <PageHeader.Breadcrumb>Contexts</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            <Tabs>
                <div className="mb-4">
                    <TabList>
                        <Tab className="nav-item">
                            <span className="nav-link text-content" data-tab-content="Your search contexts">
                                Your search contexts
                            </span>
                        </Tab>
                    </TabList>
                </div>
                <TabPanels>
                    <TabPanel>
                        <SearchContextsListTab {...props} />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </Page>
    </div>
)
