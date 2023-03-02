import { FC, HTMLAttributes, useState, useEffect } from 'react'

import { mdiInformationOutline, mdiPlus, mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Icon, Text, Tooltip } from '@sourcegraph/wildcard'

import { GetCodeHostsResult, RepositoriesResultResult } from '../../../graphql-operations'
import { ProgressBar } from '../ProgressBar'
import { FooterWidget, CustomNextButton } from '../setup-steps'
import { GET_CODE_HOSTS, GET_REPOSITORIES_BY_SERVICE } from '../remote-repositories-step/queries'

import styles from './LocalRepositoriesStep.module.scss'

// TODO: Skip button logic

interface LocalRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const LocalRepositoriesStep: FC<LocalRepositoriesStepProps> = props => {
    const { className, ...attributes } = props
    const [localRepositories, setLocalRepositories] = useState<RepositoriesResultResult>()
    const [repoPickerMode, setRepoPickerMode] = useState<string>('')

    // TODO: Edit action?

    /** TODO: Trade out for GetLocalRepositoriesByService() query once query is open
     * -->  query GetLocalRepositoriesService() {
                node (id: "SPECIAL BUILT IN SERVICE ID") {
                    ... on ExternalService {
                        id
                        kind
                        displayName
                        repoCount
                        config
                    }
                }
            }
     */
    const { data } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network',
        pollInterval: 5000,
    })
    console.log(data)

    const { data: repoData } = useQuery<RepositoriesResultResult>(GET_REPOSITORIES_BY_SERVICE, {
        fetchPolicy: 'cache-and-network',
        variables: {
            first: 25,
            externalService: 'RXh0ZXJuYWxTZXJ2aWNlOjQ5Mzc0',
        },
    })
    console.log(repoData)

    // useEffect(() => {
    //     if (!data?.externalServices.nodes) return

    //     const localRepos = useQuery<RepositoriesResultResult>(GET_REPOSITORIES_BY_SERVICE, {
    //         fetchPolicy: 'cache-and-network',
    //         variables: { id: data?.externalServices.nodes[0].id },
    //     })
    //     setLocalRepositories(localRepos.data)
    // }, [data])

    /** TODO: Implement BE file picker & local repo discovery
     * --> File picker (https://github.com/sourcegraph/sourcegraph/issues/48127)
            query GetAbsolutePath {
                getLocalAbsoluteRepositoryPath {
                    path
                }
            }
     * --> Repo discovery based on path & mode (https://github.com/sourcegraph/sourcegraph/issues/48128)
            query GetLocalRepositoriesPath($path: String!, $mode: LocalRepositoriesDiscoveryMode) {
                getDiscoveredLocalRepositories(path: $path, mode: $mode) {
                    nodes {
                        id
                        name
                        path
                    }
                }
            }
     */

    const handleRepoPicker = (value: string) => {
        setRepoPickerMode(value)
    }

    return (
        <div {...attributes} className={classNames(className)}>
            <Text className="mb-2">Add your local repositories.</Text>

            <Container>
                <ul className={styles.list}>
                    {/* {localRepositories?.length ? (
                        localRepositories.map((codeHost, index) => (
                            <li
                                key={codeHost.id}
                                className={classNames(
                                    'p-2 d-flex',
                                    styles.item,
                                    index + 1 !== localRepositories.length && styles.itemBorder
                                )}
                            >
                                <Icon svgPath={mdiGit} aria-hidden={true} className="mt-1 mr-3" />
                                <div className="d-flex flex-column">
                                    <Text weight="medium" className="mb-0">
                                        {codeHost.displayName}
                                    </Text>
                                    <Text size="small" className="text-muted mb-0">
                                        {codeHost.displayName}
                                    </Text>
                                </div>

                                // TODO: Edit here
                                <Tooltip content="Delete repository" placement="right" debounce={0}>
                                    <Button
                                        variant="secondary"
                                        className={classNames('ml-auto px-2 py-0', styles.button)}
                                    >
                                        <Icon svgPath={mdiDelete} aria-label="Delete code host connection" />
                                    </Button>
                                </Tooltip>
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
                    )} */}

                    <li className="d-flex">
                        <Button
                            onClick={() => handleRepoPicker('flat')}
                            variant="secondary"
                            className={classNames('col-md-3 col-5 mr-2 d-flex align-items-center', styles.button)}
                            outline={true}
                        >
                            <Icon svgPath={mdiPlus} aria-hidden={true} height={26} width={26} />
                            <Text weight="medium" className="ml-2 text-left mb-0">
                                Select a repository
                            </Text>
                        </Button>
                        <Button
                            onClick={() => handleRepoPicker('recursive')}
                            variant="secondary"
                            className={classNames('col-md-3 col-5 d-flex align-items-center', styles.button)}
                            outline={true}
                        >
                            <Icon svgPath={mdiPlus} aria-hidden={true} height={26} width={26} />
                            <Text weight="medium" className="ml-2 text-left mb-0">
                                Add a folder
                            </Text>
                        </Button>
                    </li>
                </ul>
            </Container>

            <FooterWidget>
                <ProgressBar />
            </FooterWidget>

            <CustomNextButton label="Skip" disabled={false} />
        </div>
    )
}
