import { type FC, useCallback, useMemo, useState, useEffect } from 'react'

import { Navigate } from 'react-router-dom'

import { RepoMetadata, type RepoMetadataItem } from '@sourcegraph/branded'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, ErrorAlert, Input, Text, LoadingSpinner, Link } from '@sourcegraph/wildcard'

import type { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import type {
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
export const RepoMetadataPage: FC<RepoMetadataPageProps> = ({
    telemetryService,
    telemetryRecorder,
    useBreadcrumb,
    repo,
    ...props
}) => {
    useBreadcrumb(BREADCRUMB)
    const [repoMetadataEnabled, status] = useFeatureFlag('repository-metadata', true)

    useEffect(() => {
        if (repoMetadataEnabled) {
            telemetryService.log('repoPage:ownershipPage:viewed')
            telemetryRecorder.recordEvent('repoPage.ownershipPage', 'viewed')
        }
    }, [repoMetadataEnabled, telemetryService, telemetryRecorder])

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
            if (
                !window.confirm(
                    `Remove metadata "${meta.key}${meta.value ? `:${meta.value}` : ''}" from this repository?`
                )
            ) {
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
                Add repository metadata to help search, filter and navigate between repositories. Repository metadata
                can also be added via the CLI and API. See the{' '}
                <Link to="/help/admin/repo/metadata">Repository Metadata Documentation</Link> to learn more.
            </Text>
            <Container className="repo-settings-metadata-page mb-2">
                {fetchError && <ErrorAlert error={fetchError} />}
                {deleteError && <ErrorAlert error={deleteError} />}
                {fetchLoading && deleteLoading && <LoadingSpinner />}
                {items.length ? (
                    <>
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
                            searchQuery.length > 0 && (
                                <Text className="text-muted m-0">No metadata containing "{searchQuery}"</Text>
                            )
                        )}
                    </>
                ) : (
                    <Text className="text-muted m-0">No metadata</Text>
                )}
            </Container>
            <AddMetadataForm onDidAdd={refetch} repoID={repo.id} />
        </Page>
    )
}
