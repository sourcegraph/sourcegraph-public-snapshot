import { useMemo } from 'react'

import { mdiPlus } from '@mdi/js'

import { useQuery } from '@sourcegraph/http-client'
import { ButtonLink, Container, Icon, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { GitHubAppsResult, GitHubAppsVariables } from '../../graphql-operations'
import { ConnectionContainer, ConnectionError, ConnectionLoading, ConnectionList } from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { GITHUB_APPS_QUERY } from './backend'
import { GitHubAppCard } from './GitHubAppCard'

import styles from './GitHubAppCard.module.scss'

export const GitHubAppsPage: React.FC = () => {
    const { data, loading, error } = useQuery<GitHubAppsResult, GitHubAppsVariables>(GITHUB_APPS_QUERY, {})
    const gitHubApps = useMemo(() => data?.gitHubApps?.nodes ?? [], [data])

    if (loading) return <LoadingSpinner />
    if (error) return <p>Error!</p>

    return (
        <>
            <PageTitle title="GitHub Apps" />
            <PageHeader
                path={[{ text: 'GitHub Apps' }]}
                className="mb-1 mt-3 test-tree-page-title"
                actions={
                    <ButtonLink to="/site-admin/github-apps/new" variant="primary" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add GitHub App
                    </ButtonLink>
                }
            />
            <Container className="mb-3 mt-3 p-3">
                <ConnectionContainer>
                    {error && <ConnectionError errors={error} />}
                    {loading && !data && <ConnectionLoading />}
                    <ConnectionList as="ul" className="list-group" aria-label="GitHub Apps">
                        {gitHubApps?.map(app => (
                            <li className={styles.listNode}>
                                <GitHubAppCard key={app.id} app={app} />
                            </li>
                        ))}
                    </ConnectionList>
                </ConnectionContainer>
            </Container>
        </>
    )
}
