import { FC, useCallback, useMemo, useState, useEffect } from 'react'

import { Navigate } from 'react-router-dom'

import { RepoMetadata, RepoMetadataItem } from '@sourcegraph/branded'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, ErrorAlert, Input, Text, LoadingSpinner, Link } from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import {
    GetRepoMetadataResult,
    GetRepoMetadataVariables,
    DeleteRepoMetadataResult,
    DeleteRepoMetadataVariables,
    RepositoryFields,
} from '../../graphql-operations'

import { AddMetadataForm } from './AddMetadataForm'
import { DELETE_REPO_METADATA_GQL, GET_REPO_METADATA_GQL } from './query'

const BREADCRUMB = { key: 'metadata', element: 'Metadata' }

interface RepoMetadataPageProps extends Pick<BreadcrumbSetters, 'useBreadcrumb'>, TelemetryProps {
    repo: Pick<RepositoryFields, 'name' | 'url' | 'metadata' | 'id'>
}

/**
 * The repository metadata page.
 */
export const RepoMetadataPage: FC<RepoMetadataPageProps> = ({ telemetryService, useBreadcrumb, repo, ...props }) => {
    useBreadcrumb(BREADCRUMB)
    const [repoMetadataEnabled, status] = useFeatureFlag('repository-metadata')

    useEffect(() => {
        if (repoMetadataEnabled) {
            telemetryService.log('repoPage:ownershipPage:viewed')
        }
    }, [repoMetadataEnabled, telemetryService])

    const {
        data,
        error: fetchError,
        refetch,
        loading: fetchLoading,
    } = useQuery<GetRepoMetadataResult, GetRepoMetadataVariables>(GET_REPO_METADATA_GQL, {
        variables: { repo: repo.name },
        pollInterval: 3000,
    })
    const items = data?.repository ? data.repository?.metadata : repo.metadata

    const [deleteRepoMetadata, { loading: deleteLoading, error: deleteError }] = useMutation<
        DeleteRepoMetadataResult,
        DeleteRepoMetadataVariables
    >(DELETE_REPO_METADATA_GQL, {})

    const onDelete = useCallback(
        (meta: RepoMetadataItem): void => {
            if (!window.confirm(`Delete metadata "${meta.key}${meta.value ? `:${meta.value}` : ''}"?`)) {
                return
            }
            deleteRepoMetadata({
                variables: {
                    repo: repo.id,
                    key: meta.key,
                },
            })
                .then(() => refetch())
                // eslint-disable-next-line no-console
                .catch(console.error)
        },
        [deleteRepoMetadata, repo.id, refetch]
    )

    const [searchQuery, setSearchQuery] = useState<string>('')
    const handleSearchChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setSearchQuery(event.target.value)
    }, [])

    const filteredMetadata = useMemo(
        () =>
            items
                .filter(({ key, value }) => {
                    const search = searchQuery.toLowerCase()
                    return key.toLowerCase().includes(search) || value?.toLowerCase().includes(search)
                })
                .map(({ key, value }) => ({ key, value })),
        [items, searchQuery]
    )

    if (status !== 'loaded') {
        return <div>Loading...</div>
    }

    if (!repoMetadataEnabled) {
        return <Navigate to={repo.url} replace={true} />
    }

    return (
        <Page>
            <PageTitle title="Repo metadata settings" />
            <PageHeader path={[{ text: 'Metadata' }]} headingElement="h2" className="mb-3" />
            <Text>
                Repository metadata allows you to search, filter and navigate between repositories. Administrators can
                add repository metadata via the web, cli or API. Learn more about{' '}
                <Link to="/help/admin/repo/metadata">Repository Metadata</Link>.
            </Text>
            <Container className="repo-settings-metadata-page mb-2">
                {fetchError && <ErrorAlert error={fetchError} />}
                {deleteError && <ErrorAlert error={deleteError} />}
                {fetchLoading && deleteLoading && <LoadingSpinner />}
                <Input
                    placeholder="Filter metadata by key or value…"
                    value={searchQuery}
                    onChange={handleSearchChange}
                    type="search"
                    className="mb-3"
                />
                {filteredMetadata.length ? (
                    <RepoMetadata items={filteredMetadata} onDelete={onDelete} />
                ) : (
                    <Text className="text-muted">No metadata containing "{searchQuery}"</Text>
                )}
            </Container>
            <AddMetadataForm onDidAdd={refetch} repoID={repo.id} />
        </Page>
    )
}
