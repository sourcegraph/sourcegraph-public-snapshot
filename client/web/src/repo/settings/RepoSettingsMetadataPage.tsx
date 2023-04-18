import { FC, useCallback, useMemo, useState, useEffect } from 'react'

import { mdiPlus } from '@mdi/js'

import { RepoMetadata } from '@sourcegraph/branded'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    Container,
    PageHeader,
    ErrorAlert,
    Input,
    Button,
    Icon,
    Modal,
    Label,
    Form,
    Alert,
    LoadingSpinner,
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

    const [addRepoMetadata, { called, loading, error, reset }] = useMutation<
        AddRepoMetadataResult,
        AddRepoMetadataVariables
    >(ADD_REPO_METADATA_GQL)

    const [isOpen, setIsOpen] = useState<boolean>(false)
    const onClose = (): void => setIsOpen(false)
    const onOpen = (): void => {
        setIsOpen(true)
        reset()
    }

    const onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        addRepoMetadata({
            variables: {
                repo: repoID,
                key,
                value,
            },
        })
            .then(() => onDidAdd())
            .catch(console.error)
    }

    return (
        <>
            <div className="mb-2 d-flex justify-content-end">
                <Button variant="primary" size="sm" onClick={onOpen}>
                    <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                    Add New Metadata
                </Button>
            </div>
            <Modal title="Add new metadata" aria-label="Add new metadata" isOpen={isOpen} onDismiss={onClose}>
                <Form onSubmit={onSubmit}>
                    {!loading && error && <ErrorAlert className="flex-grow-1 m-0 mb-3" error={error} />}
                    {!loading && !error && called && (
                        <Alert className="flex-grow-1 m-0 mb-3" variant="success">
                            Metadata added successfully
                        </Alert>
                    )}
                    <div className="form-group">
                        <Label htmlFor="metadata-key">Key*</Label>
                        <Input
                            id="metadata-key"
                            value={key}
                            onChange={handleKeyChange}
                            autoFocus={true}
                            required={true}
                            disabled={loading}
                            placeholder="e.g. 'team', 'owner', 'language'"
                        />
                    </div>
                    <div className="form-group">
                        <Label htmlFor="metadata-value">Value</Label>
                        <Input
                            id="metadata-value"
                            value={value}
                            onChange={handleValueChange}
                            disabled={loading}
                            placeholder="e.g. 'frontend', 'engineering', 'Go'"
                        />
                    </div>
                    <div className="d-flex justify-content-end">
                        <Button variant="secondary" onClick={onClose} className="mr-2" disabled={loading}>
                            Cancel
                        </Button>
                        <LoaderButton variant="primary" type="submit" loading={loading} label="Add" />
                    </div>
                </Form>
            </Modal>
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
        (key: string): void => {
            if (!window.confirm(`Delete metadata key "${key}"?`)) {
                return
            }
            deleteRepoMetadata({
                variables: {
                    repo: repo.id,
                    key,
                },
            })
                .then(() => refetch())
                .catch(console.error)
        },
        [deleteRepoMetadata, repo.id, refetch]
    )

    const [searchQuery, setSearchQuery] = useState<string>('')
    const handleSearchChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setSearchQuery(event.target.value)
    }, [])

    const filteredMetadata = useMemo(
        (): [string, string | undefined | null][] =>
            repo.keyValuePairs
                .filter(({ key, value }) => {
                    const search = searchQuery.toLowerCase()
                    return key.toLowerCase().includes(search) || value?.toLowerCase().includes(search)
                })
                .map(({ key, value }) => [key, value]),
        [repo.keyValuePairs, searchQuery]
    )

    return (
        <>
            <PageTitle title="Repo metadata settings" />
            <PageHeader path={[{ text: 'Repo metadata' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-metadata-page">
                {fetchError && <ErrorAlert error={fetchError} />}
                {deleteError && <ErrorAlert error={deleteError} />}
                {fetchLoading && deleteLoading && <LoadingSpinner />}
                <AddRepoMetadata onDidAdd={refetch} repoID={repo.id} />
                <Input
                    placeholder="Search repo metadata"
                    value={searchQuery}
                    onChange={handleSearchChange}
                    className="mb-3"
                />
                <RepoMetadata keyValuePairs={filteredMetadata} onDelete={onDelete} />
            </Container>
        </>
    )
}
