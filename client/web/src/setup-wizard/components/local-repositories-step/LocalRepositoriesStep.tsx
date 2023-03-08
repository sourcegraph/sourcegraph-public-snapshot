import { FC, HTMLAttributes, useState, useEffect } from 'react'

import { mdiInformationOutline, mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Icon, Input, Text } from '@sourcegraph/wildcard'

import { GetCodeHostsResult, ExternalServiceKind, RepositoriesResult } from '../../../graphql-operations'
import { ProgressBar } from '../ProgressBar'
import {
    GET_LOCAL_DIRECTORY_PATH,
    DISCOVER_LOCAL_REPOSITORIES,
    GET_CODE_HOSTS,
    GET_REPOSITORIES_BY_SERVICE,
} from '../remote-repositories-step/queries'
import { FooterWidget, CustomNextButton } from '../setup-steps'

import { LoaderButton } from '../../../components/LoaderButton'

import styles from './LocalRepositoriesStep.module.scss'

interface ExternalService {
    __typename: 'ExternalService'
    id: string
    kind: ExternalServiceKind
    repoCount: number
    displayName: string
    lastSyncAt: string | null
    nextSyncAt: string | null
}

interface LocalDirectoryResult {
    __typename: 'ExternalService'
    path: string
    repositories: LocalRepository[]
}

interface LocalRepository {
    name: string
    path: string
}

export interface Repository {
    __typename: 'Repository'
    id: string
    name: string
    createdAt: string
    viewerCanAdminister: boolean
    url: string
    isPrivate: boolean
    mirrorInfo: {
        __typename?: 'MirrorRepositoryInfo'
        cloned: boolean
        cloneInProgress: boolean
        updatedAt: string | null
        isCorrupted: boolean
        lastError: string | null
        byteSize: string
        shard: string | null
        corruptionLogs: { __typename?: 'RepoCorruptionLog'; timestamp: string }[]
    }
    externalRepository: { __typename?: 'ExternalRepository'; serviceType: string; serviceID: string }
}

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props
    const [localServices, setLocalServices] = useState<ExternalService>()

    const [foundRepositories, setFoundRepositories] = useState<any[]>()

    const [directoryPicker, setDirectoryPicker] = useState<boolean>(false)
    const [directoryPath, setDirectoryPath] = useState<string>('')

    const filePickerAvailable = !!window.context.localFilePickerAvailable

    // TODO: Trade out for getLocalServices() or extended externalServices(kind: "OTHER") if/when available
    const { data } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network',
        pollInterval: 5000,
    })

    /** Parse out non-hard-coded local services connected */
    useEffect(() => {
        if (!data?.externalServices.nodes) {
            return
        }

        const localServicesOnly = data.externalServices.nodes.find(
            node => node.kind === ExternalServiceKind.OTHER && node.id !== 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0'
        )
        setLocalServices(localServicesOnly)
    }, [data])

    const { data: repoData } = useQuery<RepositoriesResult>(GET_REPOSITORIES_BY_SERVICE, {
        fetchPolicy: 'cache-and-network',
        variables: {
            skip: !localServices,
            first: 25,
            externalService: localServices?.id ?? '',
        },
    })

    useEffect(() => {
        setFoundRepositories(repoData?.repositories?.nodes)
    }, [repoData])

    // TODO: Fix bug, dir picker fires on initial load
    const { data: localDirectoryPath } = useQuery<LocalDirectoryResult>(GET_LOCAL_DIRECTORY_PATH, {
        variables: {
            skip: !directoryPicker,
        },
    })

    const { data: discoveredRepositories } = useQuery<LocalDirectoryResult>(DISCOVER_LOCAL_REPOSITORIES, {
        variables: {
            skip: !directoryPath,
            dir: directoryPath,
        },
    })

    useEffect(() => {
        setFoundRepositories(discoveredRepositories?.localDirectory?.repositories)
    }, [discoveredRepositories])

    useEffect(() => {
        setDirectoryPath(localDirectoryPath?.localDirectoryPicker?.path)

        if (directoryPath) {
            setDirectoryPicker(false)
        }
    }, [localDirectoryPath])

    const handleRepoPicker = (): void => {
        if (filePickerAvailable) {
            setDirectoryPicker(true)
        }

        // TODO: Cancel edit logic
        // TODO: Fallback for non-Mac users, discover repos with simple input
    }

    const onConnect = (): void => {
        // run createExternalService(config: String!)
    }

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <>
                    <div className="d-flex w-100">
                        <Input
                            placeholder="Users/user-name/Projects/"
                            disabled={filePickerAvailable}
                            value={directoryPath || ''}
                            className="mb-0 pr-1 col-6"
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

                        {foundRepositories?.length && (
                            <Button size="sm" variant="secondary" className="ml-2">
                                Cancel
                            </Button>
                        )}
                    </div>

                    <ul className={styles.list}>
                        {/* TODO: Add loading & error state */}
                        {foundRepositories?.map((codeHost, index) => (
                            <li
                                key={codeHost.path}
                                className={classNames(
                                    'ml-3 p-2 d-flex',
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
                </>

                {foundRepositories?.length ? (
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
                            className="mr-2 mx-2"
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
