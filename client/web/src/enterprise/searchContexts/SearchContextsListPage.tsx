import React, { useState } from 'react'

import { mdiMagnify, mdiPlus } from '@mdi/js'

import { SearchContextProps } from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { PageHeader, Link, Button, Icon, Alert } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { eventLogger } from '../../tracking/eventLogger'

import { SearchContextsList } from './SearchContextsList'

import styles from './SearchContextsListPage.module.scss'

export interface SearchContextsListPageProps
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'getUserSearchContextNamespaces'>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const SearchContextsListPage: React.FunctionComponent<SearchContextsListPageProps> = ({
    authenticatedUser,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
    platformContext,
    isSourcegraphDotCom,
}) => {
    const [alert, setAlert] = useState<string | undefined>()

    return (
        <div data-testid="search-contexts-list-page" className="w-100">
            <Page>
                <PageHeader
                    actions={
                        <div className={styles.actions}>
                            <Button to="/contexts/new" variant="primary" as={Link}>
                                <Icon aria-hidden={true} svgPath={mdiPlus} />
                                Create search context
                            </Button>
                            {isSourcegraphDotCom && (
                                <Button
                                    to="https://signup.sourcegraph.com/?p=context"
                                    className="mt-2"
                                    as={Link}
                                    variant="secondary"
                                    onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                                >
                                    Search private code
                                </Button>
                            )}
                        </div>
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
                {alert && <Alert variant="danger">{alert}</Alert>}
                <div id="search-context-tabs-list" className="nav nav-tabs">
                    <div className="nav-item" role="tablist">
                        <Link
                            to="/contexts"
                            role="tab"
                            aria-selected={true}
                            aria-controls="search-context-list"
                            className="nav-link active"
                        >
                            <span className="text-content" data-tab-content="Your search contexts">
                                Available contexts
                            </span>
                        </Link>
                    </div>
                </div>
                <div role="tabpanel" id="search-context-list">
                    <SearchContextsList
                        authenticatedUser={authenticatedUser}
                        getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                        fetchSearchContexts={fetchSearchContexts}
                        platformContext={platformContext}
                        setAlert={setAlert}
                    />
                </div>
            </Page>
        </div>
    )
}
