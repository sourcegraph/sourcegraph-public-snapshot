import { FC, useCallback, useRef, useMemo, useState, useEffect } from 'react'

import { Navigate } from 'react-router-dom'
import { useDebounce } from 'use-debounce'

import { RepoMetadata, RepoMetadataItem } from '@sourcegraph/branded'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
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
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
} from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../components/Breadcrumbs'
import { LoaderButton } from '../components/LoaderButton'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import {
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    AddRepoMetadataResult,
    AddRepoMetadataVariables,
    DeleteRepoMetadataResult,
    DeleteRepoMetadataVariables,
    RepositoryFields,
    SearchRepoMetadataResult,
    SearchRepoMetadataVariables,
} from '../graphql-operations'

import {
    ADD_REPO_METADATA_GQL,
    DELETE_REPO_METADATA_GQL,
    FETCH_SETTINGS_AREA_REPOSITORY_GQL,
    SEARCH_REPO_METADATA_GQL,
} from './settings/backend'

const encodeRepoMeta = (keyOrValue: string): string => keyOrValue.replaceAll(':', '\\:')
const decodeRepoMeta = (keyOrValue: string): string => keyOrValue.replaceAll('\\:', ':')
const parseRepoMetaPair = (pair: string): [string, string] => pair.split(/(?<!\\):/g) as [string, string]

const AddRepoMetadata: FC<{ onDidAdd: () => void; repoID: string }> = ({ onDidAdd, repoID }) => {
    const [key, setKey] = useState<string>('')
    const [value, setValue] = useState<string>('')
    const valueInputRef = useRef<HTMLInputElement>(null)
    const [showSuggestions, setShowSuggestions] = useState<boolean>(false)

    const handleKeyChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setKey(event.target.value)
        setShowSuggestions(true)
    }, [])

    const handleValueChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setValue(event.target.value)
    }, [])

    const onSuggestionSelect = useCallback((suggestion: string): void => {
        const [key, value] = parseRepoMetaPair(suggestion).map(decodeRepoMeta)
        setKey(key)
        setShowSuggestions(false)
        setValue(value || '')
        valueInputRef.current?.focus()
    }, [])

    const [searchKey, searchValue] = useMemo(() => parseRepoMetaPair(key), [key])
    const [[queryKey, queryValue]] = useDebounce([searchKey, searchValue], 500)
    const { data, loading: searchByKeyLoading } = useQuery<SearchRepoMetadataResult, SearchRepoMetadataVariables>(
        SEARCH_REPO_METADATA_GQL,
        {
            variables: { queryKey: queryKey || null, queryValue: queryValue || null },
            fetchPolicy: 'network-only',
        }
    )

    const suggestions = useMemo(
        () =>
            data?.repoMetadata?.nodes.map(
                meta => encodeRepoMeta(meta.key) + (meta.value ? `:${encodeRepoMeta(meta.value)}` : '')
            ),
        [data]
    )

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

            <Container className="repo-metadata-page">
                <section>
                    <H2>Add metadata</H2>
                    <Text>Add an additional key, or key-value pair, to this repository.</Text>
                    <Form onSubmit={onSubmit}>
                        {!loading && error && <ErrorAlert className="flex-grow-1 m-0 mb-3" error={error} />}
                        <div className="form-group">
                            <Label htmlFor="metadata-key">Key</Label>
                            <Combobox onSelect={onSuggestionSelect} openOnFocus={true}>
                                <ComboboxInput
                                    id="metadata-key"
                                    required={true}
                                    disabled={loading}
                                    autoComplete="off"
                                    value={key}
                                    onChange={handleKeyChange}
                                    message="e.g. 'status', 'license', 'language'"
                                />
                                {showSuggestions && (
                                    <ComboboxPopover>
                                        {!searchByKeyLoading && (
                                            <ComboboxList>
                                                {suggestions?.map(suggestion => (
                                                    <ComboboxOption key={suggestion} value={suggestion} />
                                                ))}
                                            </ComboboxList>
                                        )}
                                    </ComboboxPopover>
                                )}
                            </Combobox>
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
                                ref={valueInputRef}
                            />
                        </div>
                        <LoaderButton variant="primary" type="submit" loading={loading} label="Add" />
                    </Form>
                </section>
            </Container>
        </>
    )
}

const BREADCRUMB = { key: 'metadata', element: 'Metadata' }

interface RepoMetadataPageProps extends Pick<BreadcrumbSetters, 'useBreadcrumb'>, TelemetryProps {
    repo: RepositoryFields
}

/**
 * The repository metadata page.
 */
export const RepoMetadataPage: FC<RepoMetadataPageProps> = ({ telemetryService, useBreadcrumb, ...props }) => {
    useBreadcrumb(BREADCRUMB)
    const [repoMetadataEnabled] = useFeatureFlag('repository-metadata')

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
            <AddRepoMetadata onDidAdd={refetch} repoID={repo.id} />
        </Page>
    )
}
