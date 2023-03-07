import { FC, HTMLAttributes, MouseEvent, useState, useEffect } from 'react'

import { mdiInformationOutline, mdiPlus, mdiGit, mdiPencil } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Icon, Input, Text, Tooltip } from '@sourcegraph/wildcard'

import { GetCodeHostsResult, ExternalServiceKind, RepositoriesResult } from '../../../graphql-operations'
import { ProgressBar } from '../ProgressBar'
import { GET_CODE_HOSTS, GET_REPOSITORIES_BY_SERVICE } from '../remote-repositories-step/queries'
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
    const [foundRepositories, setFoundRepositories] = useState<RepositoriesResult>()
    const [editLocalService, setEditLocalService] = useState<boolean>(false)

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
            externalService: localServices?.id ?? 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0',
        },
    })

    useEffect(() => {
        setFoundRepositories(repoData)
    }, [repoData])

    const handleRepoPicker = (): void => {
        if (window.context.runningOnMacOS) {
            // TODO: Implement BE file picker (getAbsolutePath()) --> https://github.com/sourcegraph/sourcegraph/issues/48127
            // onRepositoryDiscovery(,path)
        }

        setEditLocalService(true)
    }

    const onRepositoryDiscovery = (event?: MouseEvent<HTMLElement>, path?: string): void => {
        // const { data: discoveredRepos } = getDiscoveredRepositories(event?.target?.value || path)
        // setFoundRepositories(discoveredRepos)
    }

    const onConnect = (): void => {
        // run createExternalService(config: String!)
    }

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                {foundRepositories?.repositories?.nodes.length ? (
                    <>
                        {editLocalService ? (
                            <LocalRepositoryForm
                                onFind={() => onRepositoryDiscovery()}
                                onCancel={() => setEditLocalService(false)}
                            />
                        ) : (
                            <li className={classNames(styles.item, 'd-flex align-items-center p-2')}>
                                <Icon svgPath={mdiGit} aria-hidden={true} className="my-auto mr-3" />
                                <Text weight="medium" className="mb-0">
                                    Path: /User/Projects
                                </Text>

                                <Tooltip content="Edit service" placement="right" debounce={0}>
                                    <Button
                                        onClick={() => setEditLocalService(true)}
                                        variant="secondary"
                                        className={classNames('ml-auto p-2', styles.button)}
                                    >
                                        <Icon svgPath={mdiPencil} aria-label="Edit code host connection" />
                                    </Button>
                                </Tooltip>
                            </li>
                        )}

                        <ul className={styles.list}>
                            {foundRepositories?.repositories?.nodes.map((codeHost, index) => (
                                <li
                                    key={codeHost.id}
                                    className={classNames(
                                        'ml-3 p-2 d-flex',
                                        index + 1 !== foundRepositories?.repositories?.nodes.length && styles.itemBorder
                                    )}
                                >
                                    <Icon svgPath={mdiGit} aria-hidden={true} className="mt-1 mr-3" />
                                    <div className="d-flex flex-column">
                                        {/* TODO: Replace with SG relative path when available */}
                                        <Text weight="medium" className="mb-0">
                                            {codeHost.url}
                                        </Text>
                                        {/* TODO: Replace with absolute path when available */}
                                        <Text size="small" className="text-muted mb-0">
                                            {codeHost.url}
                                        </Text>
                                    </div>
                                </li>
                            ))}
                        </ul>

                        <LoaderButton
                            type="submit"
                            variant="primary"
                            className="ml-auto mr-2"
                            size="sm"
                            label="Connect"
                            onClick={onConnect}
                            alwaysShowLabel={true}
                            loading={false}
                            disabled={editLocalService}
                        />
                    </>
                ) : (
                    <>
                        {editLocalService ? (
                            <LocalRepositoryForm
                                onFind={() => onRepositoryDiscovery()}
                                onCancel={() => setEditLocalService(false)}
                            />
                        ) : (
                            <Button
                                onClick={handleRepoPicker}
                                variant="secondary"
                                className={classNames('w-100 d-flex align-items-center', styles.button)}
                                outline={true}
                            >
                                <Icon svgPath={mdiPlus} aria-hidden={true} height={26} width={26} />
                                <div className="ml-2">
                                    <Text weight="medium" className="text-left mb-0">
                                        Add existing local repositories.
                                    </Text>
                                    <Text size="small" className="text-muted text-left mb-0">
                                        Multiple folders can be selected at once.
                                    </Text>
                                </div>
                            </Button>
                        )}

                        <Text weight="bold" className="d-flex align-items-center mb-0 mt-3 font-weight-bold text-muted">
                            <Icon
                                svgPath={mdiInformationOutline}
                                className="mr-2 mx-2"
                                inline={false}
                                aria-hidden={true}
                                height={22}
                                width={22}
                            />
                            To get started, add at least one local repository to Sourcegraph.
                        </Text>
                    </>
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

interface LocalRepositoryFormProps {
    onFind: () => void
    onCancel: () => void
}

const LocalRepositoryForm: FC<LocalRepositoryFormProps> = ({ onFind, onCancel }) => {
    return (
        <div className="d-flex w-100">
            <Input label="Project path" placeholder="user/path/repo-1" className="mb-0 col-5" />

            <div className="d-flex align-items-end mb-1 col-5">
                <LoaderButton
                    type="submit"
                    variant="primary"
                    size="sm"
                    label="Find repositories"
                    onClick={onFind}
                    alwaysShowLabel={true}
                    loading={false}
                    disabled={false}
                />

                <Button size="sm" onClick={onCancel} variant="secondary" className="ml-2">
                    Cancel
                </Button>
            </div>
        </div>
    )
}
