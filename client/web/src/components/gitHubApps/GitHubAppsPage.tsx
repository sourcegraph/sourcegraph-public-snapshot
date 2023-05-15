import { useMemo } from 'react'

import { mdiPlus } from '@mdi/js'

import { useQuery } from '@sourcegraph/http-client'
import { ButtonLink, ErrorAlert, Icon, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { GitHubAppsResult, GitHubAppsVariables } from '../../graphql-operations'
import { ConnectionContainer, ConnectionLoading, ConnectionList } from '../FilteredConnection/ui'
import { PageTitle } from '../PageTitle'

import { GITHUB_APPS_QUERY } from './backend'
import { GitHubAppCard } from './GitHubAppCard'

export const GitHubAppsPage: React.FC = () => {
    const { data, loading, error, refetch } = useQuery<GitHubAppsResult, GitHubAppsVariables>(GITHUB_APPS_QUERY, {})
    const gitHubApps = useMemo(() => data?.gitHubApps?.nodes ?? [], [data])

    const reloadApps = async (): Promise<void> => {
        await refetch({})
    }

    if (loading && !data) {
        return <LoadingSpinner />
    }

    return (
        <>
            <PageTitle title="GitHub Apps" />
            <PageHeader path={[{ text: 'GitHub Apps' }]} className="mb-1" />
            <div className="d-flex align-items-center">
                Create and connect a GitHub App.
                {/* TODO: add proper link here */}
                <Link to="" className="ml-1">
                    See how GitHub App configuration works.
                </Link>
                <ButtonLink to="/site-admin/github-apps/new" className="ml-auto" variant="primary" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Create GitHub App
                </ButtonLink>
            </div>
            {error && <ErrorAlert className="mt-4 mb-0 text-left" error={error} />}
            <ConnectionContainer>
                {error && <ErrorAlert error={error} />}
                {loading && !data && <ConnectionLoading />}
                <ConnectionList as="ul" className="list-group mt-3" aria-label="GitHub Apps">
                    {!gitHubApps || gitHubApps.length === 0 ? (
                        <div className="text-center">You haven't created any GitHub Apps yet.</div>
                    ) : (
                        gitHubApps?.map(app => <GitHubAppCard key={app.id} app={app} refetch={reloadApps} />)
                    )}
                </ConnectionList>
            </ConnectionContainer>
        </>
    )
}
