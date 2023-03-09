import { FC, forwardRef, HTMLAttributes, InputHTMLAttributes, useEffect, useState } from 'react'

import { useLazyQuery } from '@apollo/client'
import { mdiGit, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'
import { parse as parseJSONC } from 'jsonc-parser'

import { ErrorLike, modify } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, ErrorAlert, Icon, Input, Text } from '@sourcegraph/wildcard'

import {
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    CodeHost,
    DiscoverLocalRepositoriesResult,
    DiscoverLocalRepositoriesVariables,
    ExternalServiceKind,
    GetCodeHostsResult,
    GetLocalDirectoryPathResult,
    UpdateRemoteCodeHostResult,
    UpdateRemoteCodeHostVariables,
} from '../../../graphql-operations'
// TODO Move these two in shared queries if they are used in different steps UI
import { ADD_CODE_HOST, GET_CODE_HOSTS, UPDATE_CODE_HOST } from '../remote-repositories-step/queries'

import { DISCOVER_LOCAL_REPOSITORIES, GET_LOCAL_DIRECTORY_PATH } from './queries'

import styles from './LocalRepositoriesStep.module.scss'

interface LocalRepositoriesStepProps extends TelemetryProps, HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { telemetryService, ...attributes } = props

    const [directoryPath, setDirectoryPath] = useState<string>('')
    const [error, setError] = useState<ErrorLike | undefined>()

    // TODO: Trade out for getLocalServices() or extended externalServices(kind: "OTHER")
    // if/when available to simplify this block
    const { data, loading } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, {
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
            const newConfig = modify(localService.config, ['root'], directoryPath)
            // We do have local service already so run update mutation
            updateLocalCodeHost({
                refetchQueries: ['GetCodeHosts'],
                variables: {
                    input: {
                        id: localService.id,
                        config: newConfig,
                        displayName: localService.displayName,
                    },
                },
            }).catch(setError)
        } else {
            // We don't have any local external service yet, so call create mutation
            addLocalCodeHost({
                refetchQueries: ['GetCodeHosts'],
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

    return (
        <div {...attributes}>
            <Text className="mb-2">Add your local repositories from your disk.</Text>

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
        </div>
    )
}

function getLocalServicePath(data?: GetCodeHostsResult): string {
    const localCodeHost = getLocalService(data)

    if (!localCodeHost) {
        return ''
    }

    const config = parseJSONC(localCodeHost.config) as Record<string, string>
    return config.root ?? ''
}

function getLocalService(data?: GetCodeHostsResult): CodeHost | null {
    if (!data) {
        return null
    }

    return (
        data.externalServices.nodes.find(
            node => node.kind === ExternalServiceKind.OTHER && node.id !== 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0'
        ) ?? null
    )
}

function createDefaultLocalServiceConfig(path: string): string {
    return `{ "url":"${window.context.srcServeGitUrl}", "root": "${path}", "repos": ["src-serve-local"] }`
}

interface LocalRepositoriesFormProps {
    isFilePickerAvailable: boolean
    error: ErrorLike | undefined
    directoryPath: string
    onDirectoryPathChange: (path: string) => void
}

const LocalRepositoriesForm: FC<LocalRepositoriesFormProps> = props => {
    const { isFilePickerAvailable, error, directoryPath, onDirectoryPathChange } = props

    const [queryPath] = useLazyQuery<GetLocalDirectoryPathResult>(GET_LOCAL_DIRECTORY_PATH, {
        fetchPolicy: 'network-only',
        onCompleted: data => data.localDirectoryPicker?.path && onDirectoryPathChange(data.localDirectoryPicker?.path),
    })

    const { data: repositoriesData } = useQuery<DiscoverLocalRepositoriesResult, DiscoverLocalRepositoriesVariables>(
        DISCOVER_LOCAL_REPOSITORIES,
        {
            skip: !directoryPath || !!error,
            fetchPolicy: 'cache-and-network',
            variables: { dir: directoryPath },
        }
    )

    const foundRepositories = repositoriesData?.localDirectory?.repositories ?? []

    return (
        <>
            <header>
                <Input
                    as={InputWitActions}
                    value={directoryPath}
                    label="Directory path"
                    disabled={isFilePickerAvailable}
                    placeholder="Users/user-name/Projects/"
                    message="You can pick git directory or folder that contains multiple git folders"
                    className={styles.filePicker}
                    onPickPath={() => queryPath()}
                    onPathReset={() => onDirectoryPathChange('')}
                />
            </header>

            {error && <ErrorAlert error={error} className="mt-3" />}
            {!error && (
                <ul className={styles.list}>
                    {foundRepositories.map(codeHost => (
                        <li key={codeHost.path} className={classNames('d-flex')}>
                            <Icon svgPath={mdiGit} aria-hidden={true} className="mt-1 mr-3" />
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

            {foundRepositories.length === 0 && !error && (
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

interface InputWitActionsProps extends InputHTMLAttributes<HTMLInputElement> {
    onPickPath: () => void
    onPathReset: () => void
}

const InputWitActions = forwardRef<HTMLInputElement, InputWitActionsProps>((props, ref) => {
    const { className, onPickPath, onPathReset, ...attributes } = props

    return (
        <div className={styles.inputRoot}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input ref={ref} {...attributes} className={classNames(className, styles.input)} />
            <Button size="sm" type="button" variant="primary" className={styles.pickPath} onClick={onPickPath}>
                Pick a path
            </Button>

            <Button size="sm" variant="secondary" className={styles.resetPath} onClick={onPathReset}>
                Reset path
            </Button>
        </div>
    )
})
