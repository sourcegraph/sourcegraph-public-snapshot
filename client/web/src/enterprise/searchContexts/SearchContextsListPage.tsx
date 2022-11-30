import React from 'react'

import { mdiMagnify, mdiPlus } from '@mdi/js'
import * as H from 'history'

import { SearchContextProps } from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { PageHeader, Link, Button, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { eventLogger } from '../../tracking/eventLogger'

import { SearchContextsList } from './SearchContextsList'

export interface SearchContextsListPageProps
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'getUserSearchContextNamespaces'>,
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
                    <>
                        <Button to="/contexts/new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                            Create search context
                        </Button>
                        {props.isSourcegraphDotCom && (
                            <Button
                                to="https://signup.sourcegraph.com/?p=context"
                                className="d-block mt-2"
                                as={Link}
                                variant="secondary"
                                onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                            >
                                Search private code
                            </Button>
                        )}
                    </>
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
                    <PageHeader.Breadcrumb icon={mdiMagnify} to="/search" aria-label="Code Search" />
                    <PageHeader.Breadcrumb>Contexts</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            <SearchContextsList {...props} />
        </Page>
    </div>
)
