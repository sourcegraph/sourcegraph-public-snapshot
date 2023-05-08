import { ChangeEvent, FC, forwardRef, HTMLAttributes, InputHTMLAttributes, useEffect, useState } from 'react'

import { useApolloClient, useLazyQuery } from '@apollo/client'
import { mdiGit } from '@mdi/js'
import classNames from 'classnames'
import { isEqual } from 'lodash'

import { ErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Alert,
    Button,
    Container,
    ErrorAlert,
    H4,
    Icon,
    Input,
    LoaderInput,
    Text,
    useDebounce,
    Tooltip,
} from '@sourcegraph/wildcard'

import {
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    DiscoverLocalRepositoriesResult,
    DiscoverLocalRepositoriesVariables,
    ExternalServiceKind,
    GetLocalDirectoryPathResult,
    GetLocalCodeHostsResult,
    DeleteRemoteCodeHostResult,
    DeleteRemoteCodeHostVariables,
} from '../../../graphql-operations'
import { ADD_CODE_HOST, DELETE_CODE_HOST } from '../../queries'
import { CodeHostExternalServiceAlert } from '../CodeHostExternalServiceAlert'
import { ProgressBar } from '../ProgressBar'
import { CustomNextButton, FooterWidget } from '../setup-steps'

import { getLocalServices, getLocalServicePaths, createDefaultLocalServiceConfig } from './helpers'
import { DISCOVER_LOCAL_REPOSITORIES, GET_LOCAL_CODE_HOSTS, GET_LOCAL_DIRECTORY_PATH } from './queries'

import styles from './LocalRepositoriesStep.module.scss'

interface LocalRepositoriesStepProps extends TelemetryProps, HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { telemetryService, ...attributes } = props

    const [directoryPaths, setDirectoryPaths] = useState<string[]>([])
    const [error, setError] = useState<ErrorLike | undefined>()

    const apolloClient = useApolloClient()

    const { data, loading } = useQuery<GetLocalCodeHostsResult>(GET_LOCAL_CODE_HOSTS, {
        fetchPolicy: 'network-only',
        // Sync local external service paths on first load
        onCompleted: data => {
            setDirectoryPaths(getLocalServicePaths(data))
        },
        onError: setError,
    })

    const [addLocalCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)
    const [deleteLocalCodeHost] = useMutation<DeleteRemoteCodeHostResult, DeleteRemoteCodeHostVariables>(
        DELETE_CODE_HOST
    )

    // Automatically creates or deletes local external service to
    // match user choosen paths for local repositories.
    useEffect(() => {
        if (loading) {
            return
        }

        setError(undefined)

        const localServices = getLocalServices(data)
        const localServicePaths = getLocalServicePaths(data)
        const havePathsChanged = !isEqual(directoryPaths, localServicePaths)

        // Do nothing if paths haven't changed
        if (!havePathsChanged) {
            return
        }

        async function syncExternalServices(): Promise<void> {
            // Create/update local external services
            for (const directoryPath of directoryPaths) {
                // If we already have a local external service for this path, skip it
                if (localServicePaths.includes(directoryPath)) {
                    continue
                }

                // Create a new local external service for this path
                await addLocalCodeHost({
                    variables: {
                        input: {
                            displayName: `Local repositories service (${directoryPath})`,
                            config: createDefaultLocalServiceConfig(directoryPath),
                            kind: ExternalServiceKind.OTHER,
                        },
                    },
                })
            }

            // Delete local external services that are no longer in the list
            for (const localService of localServices || []) {
                // If we still have a local external service for this path, skip it
                if (directoryPaths.includes(localService.path)) {
                    continue
                }

                // Delete local external service for this path
                await deleteLocalCodeHost({
                    variables: {
                        id: localService.id,
                    },
                })
            }

            // Refetch local external services and status after all mutations have been completed.
            await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts', 'StatusAndRepoStats'] })
        }

        syncExternalServices().catch(setError)
    }, [directoryPaths, data, loading, addLocalCodeHost, deleteLocalCodeHost, apolloClient])

    useEffect(() => {
        telemetryService.log('SetupWizardLandedAddLocalCode')
    }, [telemetryService])

    const handleNextButtonClick = (): void => {
        if (!directoryPaths) {
            telemetryService.log('SetupWizardSkippedAddLocalCode')
        }
    }

    // Try to find autogenerated local external service to list built-in repositories
    // below manually added repositories section
    const autogeneratedServicePaths = getLocalServices(data, true).map(item => item.path)

    return (
        <div {...attributes}>
            <Text className="mb-2">Add your local repositories from your disk.</Text>

            <CodeHostExternalServiceAlert />

            <Container className={styles.content}>
                {!loading && (
                    <LocalRepositoriesForm
                        isFilePickerAvailable={window.context.localFilePickerAvailable}
                        error={error}
                        directoryPaths={directoryPaths}
                        onDirectoryPathsChange={setDirectoryPaths}
                    />
                )}

                {!loading && autogeneratedServicePaths.length > 0 && (
                    <BuiltInRepositories directoryPaths={autogeneratedServicePaths} />
                )}
            </Container>

            <FooterWidget>
                <ProgressBar />
            </FooterWidget>

            <CustomNextButton
                label={directoryPaths.length > 0 ? 'Next' : 'Skip'}
                tooltip={directoryPaths.length === 0 ? 'You can get back to this step later' : ''}
                onClick={handleNextButtonClick}
            />
        </div>
    )
}

interface LocalRepositoriesFormProps {
    isFilePickerAvailable: boolean
    error: ErrorLike | undefined
    directoryPaths: string[]
    onDirectoryPathsChange: (paths: string[]) => void
}

const LocalRepositoriesForm: FC<LocalRepositoriesFormProps> = props => {
    const { isFilePickerAvailable, error, directoryPaths, onDirectoryPathsChange } = props

    const [internalPaths, setInternalPaths] = useState(directoryPaths)
    const [queryPath] = useLazyQuery<GetLocalDirectoryPathResult>(GET_LOCAL_DIRECTORY_PATH, {
        fetchPolicy: 'network-only',
        onCompleted: data =>
            data.localDirectoriesPicker?.paths && onDirectoryPathsChange(data.localDirectoriesPicker?.paths),
    })

    const { data: repositoriesData, loading } = useQuery<
        DiscoverLocalRepositoriesResult,
        DiscoverLocalRepositoriesVariables
    >(DISCOVER_LOCAL_REPOSITORIES, {
        skip: directoryPaths.length <= 0 || !!error,
        fetchPolicy: 'cache-and-network',
        variables: { paths: directoryPaths },
    })

    // By default, input is disabled so this callback won't be fired
    // but in case if backend-based file picker isn't supported in OS
    // that is running sg instance we fall back on common input where user
    // should file path manually
    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setInternalPaths(event.target.value.split(','))
    }

    const handlePathReset = (): void => {
        setInternalPaths([])
        onDirectoryPathsChange([])
    }

    const debouncedInternalPaths = useDebounce(internalPaths, 1000)

    // Sync internal state with parent logic
    useEffect(() => {
        onDirectoryPathsChange(debouncedInternalPaths)
    }, [debouncedInternalPaths, onDirectoryPathsChange])

    // Use internal path only if backend-based file picker is unavailable
    const paths = isFilePickerAvailable ? directoryPaths : internalPaths
    const initialState = !repositoriesData && !error && !loading
    const foundRepositories = repositoriesData?.localDirectories?.repositories ?? []
    const zeroResultState =
        paths.length > 0 && !error && repositoriesData && repositoriesData.localDirectories.repositories.length === 0

    return (
        <>
            <header>
                <Input
                    as={InputWithActions}
                    value={paths.join(', ')}
                    label="Directory path"
                    isFilePickerMode={isFilePickerAvailable}
                    placeholder="/Users/user-name/Projects/"
                    message="Pick a git directory or folder that contains multiple git folders"
                    isProcessing={loading}
                    className={styles.filePicker}
                    // eslint-disable-next-line @typescript-eslint/no-misused-promises
                    onPickPath={() => queryPath()}
                    onPathReset={handlePathReset}
                    onChange={handleInputChange}
                />
            </header>

            {error && <ErrorAlert error={error} className="mt-3" />}

            {!error && (
                <ul className={styles.list}>
                    {foundRepositories.map(codeHost => (
                        <li key={codeHost.path} className={classNames('d-flex')}>
                            <Icon svgPath={mdiGit} size="md" aria-hidden={true} className="mt-1 mr-3" />
                            <div className="d-flex flex-column">
                                <Text weight="medium" className="mb-0">
                                    {codeHost.name}
                                </Text>
                                <Text size="small" className="text-muted mb-0">
                                    {codeHost.path}
                                </Text>
                            </div>
                        </li>
                    ))}
                </ul>
            )}

            {zeroResultState && (
                <Alert variant="primary" className="mt-3 mb-0">
                    <H4>We couldn't resolve any git repositories by the current path</H4>
                    Try to use different path that contains .git repositories
                </Alert>
            )}

            {initialState && (
                <Alert variant="secondary" className="mt-3 mb-0">
                    <Text className="mb-0 text-muted">
                        Pick a path to see a list of local repositories that you want to have in the Sourcegraph App
                    </Text>
                </Alert>
            )}
        </>
    )
}

interface InputWithActionsProps extends InputHTMLAttributes<HTMLInputElement> {
    isFilePickerMode: boolean
    isProcessing: boolean
    onPickPath: () => void
    onPathReset: () => void
}

/**
 * Renders either file picker input (non-editable but clickable and with "pick a path" action button or
 * simple input where user can input path manually.
 */
const InputWithActions = forwardRef<HTMLInputElement, InputWithActionsProps>(function InputWithActions(props, ref) {
    const { isFilePickerMode, isProcessing, onPickPath, onPathReset, className, value, ...attributes } = props

    return (
        <div className={styles.inputRoot}>
            <Tooltip content={isFilePickerMode ? value : undefined}>
                <LoaderInput
                    loading={isProcessing}
                    className={styles.inputLoader}
                    onClick={isFilePickerMode ? onPickPath : undefined}
                >
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <input
                        {...attributes}
                        ref={ref}
                        value={value}
                        disabled={isFilePickerMode}
                        className={classNames(className, styles.input, { [styles.inputWithAction]: isFilePickerMode })}
                    />
                </LoaderInput>
            </Tooltip>
            {isFilePickerMode && (
                <Button size="sm" type="button" variant="primary" className={styles.pickPath} onClick={onPickPath}>
                    Pick a path
                </Button>
            )}

            <Button size="sm" variant="secondary" className={styles.resetPath} onClick={onPathReset}>
                Reset path
            </Button>
        </div>
    )
})

interface BuiltInRepositoriesProps {
    directoryPaths: string[]
}

const BuiltInRepositories: FC<BuiltInRepositoriesProps> = props => {
    const { directoryPaths } = props

    const { data: repositoriesData, loading } = useQuery<
        DiscoverLocalRepositoriesResult,
        DiscoverLocalRepositoriesVariables
    >(DISCOVER_LOCAL_REPOSITORIES, {
        variables: { paths: directoryPaths },
    })

    const totalNumberOfRepositories = repositoriesData?.localDirectories.repositories.length ?? 0

    if (loading || !repositoriesData || totalNumberOfRepositories === 0) {
        return null
    }

    const foundRepositories = repositoriesData.localDirectories.repositories

    return (
        <section className="mt-4">
            <hr />
            <H4 className="mt-3 mb-1">Built-in repositories</H4>
            <Text size="small" className="text-muted">
                You're running the Sourcegraph app from your terminal. We found the repositories below in your path.
            </Text>
            <ul className={styles.list}>
                {foundRepositories.map(repository => (
                    <li key={repository.path} className={classNames('d-flex')}>
                        <Icon svgPath={mdiGit} size="md" aria-hidden={true} className="mt-1 mr-3" />
                        <div className="d-flex flex-column">
                            <Text weight="medium" className="mb-0">
                                {repository.name}
                            </Text>
                            <Text size="small" className="text-muted mb-0">
                                {repository.path}
                            </Text>
                        </div>
                    </li>
                ))}
            </ul>
        </section>
    )
}
