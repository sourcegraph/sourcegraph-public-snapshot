import { ChangeEvent, FC, forwardRef, HTMLAttributes, InputHTMLAttributes, useEffect, useState } from 'react'

import { useLazyQuery } from '@apollo/client'
import { mdiGit, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

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
} from '@sourcegraph/wildcard'

import {
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    DiscoverLocalRepositoriesResult,
    DiscoverLocalRepositoriesVariables,
    ExternalServiceKind,
    GetLocalDirectoryPathResult,
    UpdateRemoteCodeHostResult,
    UpdateRemoteCodeHostVariables,
    GetLocalCodeHostsResult,
} from '../../../graphql-operations'
import { ADD_CODE_HOST, UPDATE_CODE_HOST } from '../../queries'
import { CodeHostExternalServiceAlert } from '../CodeHostExternalServiceAlert'
import { ProgressBar } from '../ProgressBar'
import { CustomNextButton, FooterWidget } from '../setup-steps'

import { getLocalService, getLocalServicePath, createDefaultLocalServiceConfig } from './helpers'
import { DISCOVER_LOCAL_REPOSITORIES, GET_LOCAL_CODE_HOSTS, GET_LOCAL_DIRECTORY_PATH } from './queries'

import styles from './LocalRepositoriesStep.module.scss'

interface LocalRepositoriesStepProps extends TelemetryProps, HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { telemetryService, ...attributes } = props

    const [directoryPath, setDirectoryPath] = useState<string>('')
    const [error, setError] = useState<ErrorLike | undefined>()

    // TODO: Trade out for getLocalServices() or extended externalServices(kind: "OTHER")
    // if/when available to simplify this block
    const { data, loading } = useQuery<GetLocalCodeHostsResult>(GET_LOCAL_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network',
        // Sync local state and local external service path
        onCompleted: data => setDirectoryPath(getLocalServicePath(data)),
        onError: setError,
    })
    const [addLocalCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)
    const [updateLocalCodeHost] = useMutation<UpdateRemoteCodeHostResult, UpdateRemoteCodeHostVariables>(
        UPDATE_CODE_HOST
    )

    // Automatically create or update (if local service already exists) a local
    // external service as user changes the absolute path for local repositories
    useEffect(() => {
        if (loading) {
            return
        }

        setError(undefined)

        const localService = getLocalService(data)
        const localServicePath = getLocalServicePath(data)
        const hasPathChanged = localServicePath !== directoryPath

        // Do nothing if path hasn't changed
        if (!hasPathChanged) {
            return
        }

        if (localService) {
            // We do have local service already so run update mutation
            updateLocalCodeHost({
                refetchQueries: ['GetLocalCodeHosts', 'StatusAndRepoStats'],
                variables: {
                    input: {
                        id: localService.id,
                        config: createDefaultLocalServiceConfig(directoryPath),
                        displayName: 'Local repositories service',
                    },
                },
            }).catch(setError)
        } else {
            // We don't have any local external service yet, so call create mutation
            addLocalCodeHost({
                refetchQueries: ['GetLocalCodeHosts', 'StatusAndRepoStats'],
                variables: {
                    input: {
                        displayName: 'Local repositories service',
                        config: createDefaultLocalServiceConfig(directoryPath),
                        kind: ExternalServiceKind.OTHER,
                    },
                },
            }).catch(setError)
        }
    }, [directoryPath, data, loading, addLocalCodeHost, updateLocalCodeHost])

    useEffect(() => {
        telemetryService.log('SetupWizardLandedAddLocalCode')
    }, [telemetryService])

    const handleNextButtonClick = (): void => {
        if (!directoryPath) {
            telemetryService.log('SetupWizardSkippedAddLocalCode')
        }
    }

    return (
        <div {...attributes}>
            <Text className="mb-2">Add your local repositories from your disk.</Text>

            <CodeHostExternalServiceAlert />

            <Container className={styles.content}>
                {!loading && (
                    <LocalRepositoriesForm
                        isFilePickerAvailable={window.context.localFilePickerAvailable}
                        error={error}
                        directoryPath={directoryPath}
                        onDirectoryPathChange={setDirectoryPath}
                    />
                )}
            </Container>

            <FooterWidget>
                <ProgressBar />
            </FooterWidget>

            <CustomNextButton
                label={directoryPath ? 'Next' : 'Skip'}
                tooltip={!directoryPath ? 'You can get back to this step later' : ''}
                onClick={handleNextButtonClick}
            />
        </div>
    )
}

interface LocalRepositoriesFormProps {
    isFilePickerAvailable: boolean
    error: ErrorLike | undefined
    directoryPath: string
    onDirectoryPathChange: (path: string) => void
}

const LocalRepositoriesForm: FC<LocalRepositoriesFormProps> = props => {
    const { isFilePickerAvailable, error, directoryPath, onDirectoryPathChange } = props

    const [internalPath, setInternalPath] = useState(directoryPath)
    const [queryPath] = useLazyQuery<GetLocalDirectoryPathResult>(GET_LOCAL_DIRECTORY_PATH, {
        fetchPolicy: 'network-only',
        onCompleted: data => data.localDirectoryPicker?.path && onDirectoryPathChange(data.localDirectoryPicker?.path),
    })

    const { data: repositoriesData, loading } = useQuery<
        DiscoverLocalRepositoriesResult,
        DiscoverLocalRepositoriesVariables
    >(DISCOVER_LOCAL_REPOSITORIES, {
        skip: !directoryPath || !!error,
        fetchPolicy: 'cache-and-network',
        variables: { dir: directoryPath },
    })

    // By default, input is disabled so this callback won't be fired
    // but in case if backend-based file picker isn't supported in OS
    // that is running sg instance we fall back on common input where user
    // should file path manually
    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setInternalPath(event.target.value)
    }

    const handlePathReset = (): void => {
        setInternalPath('')
        onDirectoryPathChange('')
    }

    const debouncedInternalPath = useDebounce(internalPath, 1000)

    // Sync internal state with parent logic
    useEffect(() => {
        onDirectoryPathChange(debouncedInternalPath)
    }, [debouncedInternalPath, onDirectoryPathChange])

    // Use internal path only if backend-based file picker is unavailable
    const path = isFilePickerAvailable ? directoryPath : internalPath
    const initialState = !repositoriesData && !error && !loading
    const foundRepositories = repositoriesData?.localDirectory?.repositories ?? []
    const zeroResultState =
        path && !error && repositoriesData && repositoriesData.localDirectory.repositories.length === 0

    return (
        <>
            <header>
                <Input
                    as={InputWitActions}
                    value={path}
                    label="Directory path"
                    disabled={isFilePickerAvailable}
                    placeholder="/Users/user-name/Projects/"
                    message="Pick a git directory or folder that contains multiple git folders"
                    isProcessing={loading}
                    className={styles.filePicker}
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
                <Alert variant="primary" className="mt-3 ">
                    <H4>We couldn't resolve any git repositories by the current path</H4>
                    Try to use different path that contains .git repositories
                </Alert>
            )}

            {initialState && (
                <Text className="d-flex align-items-center mb-0 mt-3 text-muted">
                    <Icon
                        svgPath={mdiInformationOutline}
                        className="mr-2 mx-1"
                        inline={false}
                        aria-hidden={true}
                        height={22}
                        width={22}
                    />
                    Pick a path to see a list of local repositories that you want to have in the Sourcegraph App.
                </Text>
            )}
        </>
    )
}

interface InputWithActionsProps extends InputHTMLAttributes<HTMLInputElement> {
    isProcessing: boolean
    onPickPath: () => void
    onPathReset: () => void
}

const InputWitActions = forwardRef<HTMLInputElement, InputWithActionsProps>((props, ref) => {
    const { isProcessing, onPickPath, onPathReset, className, disabled, ...attributes } = props

    return (
        <div className={styles.inputRoot}>
            <LoaderInput loading={isProcessing} className="flex-grow-1">
                {/* eslint-disable-next-line react/forbid-elements */}
                <input
                    {...attributes}
                    ref={ref}
                    disabled={disabled}
                    className={classNames(className, styles.input, { [styles.inputWithAction]: disabled })}
                />
            </LoaderInput>

            {disabled && (
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
