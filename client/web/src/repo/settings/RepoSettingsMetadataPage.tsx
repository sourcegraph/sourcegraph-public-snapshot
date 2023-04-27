import { FC, useCallback, useMemo, useState } from 'react'

import { RepoMetadata, RepoMetadataItem } from '@sourcegraph/branded'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    Container,
    PageHeader,
    ErrorAlert,
    Input,
    Text,
    Label,
    Form,
    Alert,
    LoadingSpinner,
    H2,
    Link,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { PageTitle } from '../../components/PageTitle'
import {
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    AddRepoMetadataResult,
    AddRepoMetadataVariables,
    DeleteRepoMetadataResult,
    DeleteRepoMetadataVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { ADD_REPO_METADATA_GQL, DELETE_REPO_METADATA_GQL, FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'

const AddRepoMetadata: FC<{ onDidAdd: () => void; repoID: string }> = ({ onDidAdd, repoID }) => {
    const [key, setKey] = useState<string>('')
    const [value, setValue] = useState<string>('')

    const handleKeyChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setKey(event.target.value)
    }, [])

    const handleValueChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setValue(event.target.value)
    }, [])

    const [addRepoMetadata, { called, loading, error }] = useMutation<AddRepoMetadataResult, AddRepoMetadataVariables>(
        ADD_REPO_METADATA_GQL
    )

    const onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        addRepoMetadata({
            variables: {
                repo: repoID,
                key,
                value,
            },
        })
            .then(() => {
                onDidAdd()
                setKey('')
                setValue('')
            })
            // eslint-disable-next-line no-console
            .catch(console.error)
    }

    return (
        <>
            {!loading && !error && called && (
                <Alert className="flex-grow-1 m-0 mb-3" variant="success">
                    Metadata added
                </Alert>
            )}

            <Container className="repo-settings-metadata-page">
                <section>
                    <H2>Add metadata</H2>
                    <Text>Add an additional key, or key-value pair, to this repository.</Text>
                    <Form onSubmit={onSubmit}>
                        {!loading && error && <ErrorAlert className="flex-grow-1 m-0 mb-3" error={error} />}

                        <div className="form-group">
                            <Label htmlFor="metadata-key">Key</Label>
                            <Input
                                id="metadata-key"
                                value={key}
                                onChange={handleKeyChange}
                                autoFocus={true}
                                autoComplete="off"
                                required={true}
                                disabled={loading}
                                message="e.g. 'status', 'license', 'language'"
                            />
                        </div>
                        <div className="form-group">
                            <Label htmlFor="metadata-value">Value (optional)</Label>
                            <Input
                                id="metadata-value"
                                value={value}
                                autoComplete="off"
                                onChange={handleValueChange}
                                disabled={loading}
                                message="e.g. 'deprecated', 'MIT', 'Go'"
                            />
                        </div>
                        <LoaderButton variant="primary" type="submit" loading={loading} label="Add" />
                    </Form>
                </section>
            </Container>
        </>
    )
}

interface RepoSettingsMetadataPageProps {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings metadata page.
 */
export const RepoSettingsMetadataPage: FC<RepoSettingsMetadataPageProps> = props => {
    eventLogger.logPageView('RepoSettingsMetadata')
    const {
        data,
        error: fetchError,
        refetch,
        loading: fetchLoading,
    } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(FETCH_SETTINGS_AREA_REPOSITORY_GQL, {
        variables: { name: props.repo.name },
        pollInterval: 3000,
    })
    const repo = data?.repository ? data.repository : props.repo

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
            repo.metadata
                .filter(({ key, value }) => {
                    const search = searchQuery.toLowerCase()
                    return key.toLowerCase().includes(search) || value?.toLowerCase().includes(search)
                })
                .map(({ key, value }) => ({ key, value })),
        [repo.metadata, searchQuery]
    )

    return (
        <>
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
            <AddRepoMetadata onDidAdd={refetch} repoID={repo.id} />
        </>
    )
}
