import { FC, HTMLAttributes, useState, useEffect } from 'react'

import { mdiInformationOutline, mdiPlus, mdiGit, mdiPencil } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Icon, Text, Tooltip } from '@sourcegraph/wildcard'

import { GetCodeHostsResult, ExternalServiceKind, RepositoriesResult } from '../../../graphql-operations'
import { ProgressBar } from '../ProgressBar'
import { FooterWidget, CustomNextButton } from '../setup-steps'
import { LocalRepositoryForm } from './components/LocalRepositoryForm'
import { GET_CODE_HOSTS, GET_REPOSITORIES_BY_SERVICE } from '../remote-repositories-step/queries'

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
        corruptionLogs: Array<{ __typename?: 'RepoCorruptionLog'; timestamp: string }>
    }
    externalRepository: { __typename?: 'ExternalRepository'; serviceType: string; serviceID: string }
}

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props
    const [localServices, setLocalServices] = useState<ExternalService[]>()
    const [newRepositoryForm, setNewRepositoryForm] = useState<boolean>(false)
    const [repositoryInEdit, setRepositoryInEdit] = useState<Repository | null>(null)

    // TODO: Trade out for getLocalServices() or extended externalServices(kind: "OTHER") if/when available
    const { data } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network',
        pollInterval: 5000,
    })

    useEffect(() => {
        if (!data?.externalServices.nodes) return

        const localServicesOnly = data.externalServices.nodes.filter(node => node.kind === ExternalServiceKind.OTHER)
        setLocalServices(localServicesOnly)
    }, [data])

    // TODO: Map through localServices
    const { data: repoData } = useQuery<RepositoriesResult>(GET_REPOSITORIES_BY_SERVICE, {
        fetchPolicy: 'cache-and-network',
        variables: {
            skip: !localServices,
            first: 25,
            externalService: 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0',
        },
    })
    console.log(repoData)

    const handleRepoPicker = () => {
        if (window.context.runningOnMacOS) {
            // TODO: Implement BE file picker (getAbsolutePath()) --> https://github.com/sourcegraph/sourcegraph/issues/48127
        }

        // TODO: Populate form input
        setNewRepositoryForm(true)
    }
    // TODO: Implement local repo discovery (getDiscoveredLocalRepos()) --> https://github.com/sourcegraph/sourcegraph/issues/48128

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <ul className={styles.list}>
                    {repoData?.repositories?.nodes.length ? (
                        repoData?.repositories?.nodes.map((codeHost, index) => (
                            <li
                                key={codeHost.id}
                                className={classNames(
                                    'p-2 d-flex',
                                    styles.item,
                                    index + 1 !== repoData?.repositories?.nodes.length && styles.itemBorder
                                )}
                            >
                                {codeHost.id === repositoryInEdit?.id ? (
                                    <LocalRepositoryForm
                                        repositoryInEdit={repositoryInEdit}
                                        onCancel={() => setRepositoryInEdit(null)}
                                    />
                                ) : (
                                    <>
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

                                        <Tooltip content="Edit repository" placement="right" debounce={0}>
                                            <Button
                                                onClick={() => setRepositoryInEdit(codeHost)}
                                                disabled={!!repositoryInEdit}
                                                variant="secondary"
                                                className={classNames('ml-auto px-2 py-0', styles.button)}
                                            >
                                                <Icon svgPath={mdiPencil} aria-label="Edit code host connection" />
                                            </Button>
                                        </Tooltip>
                                    </>
                                )}
                            </li>
                        ))
                    ) : (
                        <Text weight="bold" className="d-flex align-items-center font-weight-bold text-muted">
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
                    )}

                    <li>
                        {newRepositoryForm ? (
                            <LocalRepositoryForm onCancel={() => setNewRepositoryForm(false)} />
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
                    </li>
                </ul>
            </Container>

            <FooterWidget>
                <ProgressBar />
            </FooterWidget>

            {/* TODO: Skip button logic */}
            <CustomNextButton label="Skip" disabled={false} />
        </div>
    )
}
