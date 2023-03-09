import { FC, HTMLAttributes, useState, useEffect } from 'react'

import { useLazyQuery } from '@apollo/client'
import { mdiInformationOutline, mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { useQuery, useMutation } from '@sourcegraph/http-client'
import { Button, Container, Icon, Input, Text } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import {
    GetCodeHostsResult,
    ExternalServiceKind,
    RepositoriesResult,
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    CodeHost,
    GetLocalDirectoryPathResult,
    DiscoverLocalRepositoriesResult,
    DiscoverLocalRepositoriesVariables,
} from '../../../graphql-operations'
import { ProgressBar } from '../ProgressBar'
// TODO Move these two in shared queries if they are used in different steps UI
import { GET_CODE_HOSTS, ADD_CODE_HOST } from '../remote-repositories-step/queries'
import { FooterWidget, CustomNextButton } from '../setup-steps'

import { GET_LOCAL_DIRECTORY_PATH, DISCOVER_LOCAL_REPOSITORIES, GET_REPOSITORIES_BY_SERVICE } from './queries'

import styles from './LocalRepositoriesStep.module.scss'

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props

    // TODO Move this to the props
    const filePickerAvailable = !!window.context.localFilePickerAvailable

    /** Parse out non-hard-coded local services connected based off GET_CODE_HOSTS + GET_REPOSITORIES_BY_SERVICE */
    const [foundRepositories, setFoundRepositories] = useState<any[]>()
    const [directoryPicker, setDirectoryPicker] = useState<boolean>(false)
    const [directoryPath, setDirectoryPath] = useState<string>('')

    const [addRemoteCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)

    // TODO: Trade out for getLocalServices() or extended externalServices(kind: "OTHER") if/when available to simplify this block
    const { data } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, { fetchPolicy: 'cache-and-network' })

    // Filter out common services and get only non-generated local "Other" service
    const localService = getLocalService(data)

    const { data: repoData } = useQuery<RepositoriesResult>(GET_REPOSITORIES_BY_SERVICE, {
        skip: !localService,
        fetchPolicy: 'cache-and-network',
        variables: {
            first: 25,
            externalService: localService?.id ?? '',
        },
    })

    // Updates with any previously synced non-automated local services
    useEffect(() => {
        setFoundRepositories(repoData?.repositories?.nodes)
    }, [repoData])

    // TODO: Fix bug, dir picker fires on initial load
    const { data: localDirectoryPath } = useQuery<GetLocalDirectoryPathResult>(GET_LOCAL_DIRECTORY_PATH, {
        skip: !directoryPicker,
    })

    const { data: discoveredRepositories } = useQuery<DiscoverLocalRepositoriesResult>(DISCOVER_LOCAL_REPOSITORIES, {
        skip: !directoryPath,
        variables: { dir: directoryPath },
    })

    // Updates discovery repo list based off given directory path
    useEffect(() => {
        setFoundRepositories(discoveredRepositories?.localDirectory?.repositories)
    }, [discoveredRepositories])

    useEffect(() => {
        setDirectoryPath(localDirectoryPath?.localDirectoryPicker?.path ?? '')

        if (directoryPath) {
            setDirectoryPicker(false)
        }
    }, [directoryPath, localDirectoryPath])

    const handleRepoPicker = (): void => {
        if (filePickerAvailable) {
            setDirectoryPicker(true)
        } else {
            // TODO: Fallback for non-Mac users, discover repos with simple input
        }
    }

    const onConnect = async (): Promise<void> => {
        await addRemoteCodeHost({
            variables: {
                input: {
                    kind: ExternalServiceKind.OTHER,
                    // TODO: setup config & jsonify
                    displayName: '',
                    config: `{
                        url: '',
                        repos: [],
                    }`,
                },
            },
        })
    }

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <div className="d-flex w-100">
                    <Input
                        placeholder="Users/user-name/Projects/"
                        disabled={filePickerAvailable}
                        value={directoryPath || ''}
                        className="mb-0 pl-0 pr-1 col-6"
                    />

                    <LoaderButton
                        type="submit"
                        variant="primary"
                        size="sm"
                        label="Pick a path"
                        onClick={() => handleRepoPicker()}
                        alwaysShowLabel={true}
                        loading={false}
                        disabled={false}
                    />

                    {/* TODO: Only show cancel if no local repos are saved. Enabled cancel action */}
                    {foundRepositories?.length && (
                        <Button size="sm" variant="secondary" className="ml-2">
                            Cancel
                        </Button>
                    )}
                </div>

                <ul className={styles.list}>
                    {/* TODO: Add loading & error state for discovery */}
                    {foundRepositories?.map((codeHost, index) => (
                        <li
                            key={codeHost.path}
                            className={classNames(
                                'p-2 d-flex',
                                index + 1 !== foundRepositories?.length && styles.itemBorder
                            )}
                        >
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

                {foundRepositories?.length ? (
                    // TODO: Add loading & error state for adding local service call
                    // TODO: If this service is already saved, `CONNECT` changes to `UPDATE` with a sibling `DELETE` btn/action
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        className="ml-auto mr-2"
                        size="sm"
                        label="Connect"
                        onClick={onConnect}
                        alwaysShowLabel={true}
                        loading={false}
                        disabled={false}
                    />
                ) : (
                    <Text weight="bold" className="d-flex align-items-center mb-0 mt-3 font-weight-bold text-muted">
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
            </Container>

            <FooterWidget>
                <ProgressBar />
            </FooterWidget>

            {/* TODO: Skip button logic */}
            <CustomNextButton label="Skip" disabled={false} />
        </div>
    )
}

interface LocalRepositoriesFormProps {
    isFilePickerAvailable: boolean
    directoryPath?: string
    onSubmit: (path: string) => void
}

const LocalRepositoriesForm: FC<LocalRepositoriesFormProps> = props => {
    const { isFilePickerAvailable, directoryPath = '', onSubmit } = props

    const [path, setPath] = useState<string>(directoryPath)

    const [queryPath] = useLazyQuery<GetLocalDirectoryPathResult>(GET_LOCAL_DIRECTORY_PATH, {
        fetchPolicy: 'network-only',
        onCompleted: data => data.localDirectoryPicker?.path && setPath(data.localDirectoryPicker?.path),
    })

    const { data: repositoriesData } = useQuery<DiscoverLocalRepositoriesResult, DiscoverLocalRepositoriesVariables>(
        DISCOVER_LOCAL_REPOSITORIES,
        {
            skip: !path,
            fetchPolicy: 'cache-and-network',
            variables: { dir: path },
        }
    )

    const foundRepositories = repositoriesData?.localDirectory?.repositories ?? []

    return (
        <section>
            <div className="d-flex w-100">
                <Input
                    value={path}
                    disabled={isFilePickerAvailable}
                    placeholder="Users/user-name/Projects/"
                    className="mb-0 pl-0 pr-1 col-6"
                />

                <LoaderButton
                    size="sm"
                    type="button"
                    variant="primary"
                    label="Pick a path"
                    loading={false}
                    disabled={false}
                    alwaysShowLabel={true}
                    onClick={() => queryPath()}
                />

                {path.length > 0 && (
                    <Button size="sm" variant="secondary" className="ml-2">
                        Reset path
                    </Button>
                )}
            </div>

            <ul className={styles.list}>
                {/* TODO: Add loading & error state for discovery */}
                {foundRepositories?.map((codeHost, index) => (
                    <li
                        key={codeHost.path}
                        className={classNames(
                            'p-2 d-flex',
                            index + 1 !== foundRepositories?.length && styles.itemBorder
                        )}
                    >
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

            {foundRepositories?.length ? (
                // TODO: Add loading & error state for adding local service call
                // TODO: If this service is already saved, `CONNECT` changes to `UPDATE` with a sibling `DELETE` btn/action
                <LoaderButton
                    type="submit"
                    variant="primary"
                    className="ml-auto mr-2"
                    size="sm"
                    label="Connect"
                    alwaysShowLabel={true}
                    onClick={() => onSubmit(path)}
                />
            ) : (
                <Text weight="bold" className="d-flex align-items-center mb-0 mt-3 font-weight-bold text-muted">
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
        </section>
    )
}

function getLocalService(data?: GetCodeHostsResult): CodeHost | null {
    if (!data?.externalServices.nodes) {
        return null
    }

    return (
        data.externalServices.nodes.find(
            node => node.kind === ExternalServiceKind.OTHER && node.id !== 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0'
        ) ?? null
    )
}
