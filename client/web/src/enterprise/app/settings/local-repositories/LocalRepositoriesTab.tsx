import { FC, ReactNode, useLayoutEffect, useMemo } from 'react'

import { mdiFolderMultipleOutline, mdiFolderMultiplePlusOutline, mdiGit, mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import { pluralize } from '@sourcegraph/common'
import {
    Text,
    Icon,
    Button,
    Container,
    PageHeader,
    Collapse,
    CollapsePanel,
    LoadingSpinner,
    ErrorAlert,
} from '@sourcegraph/wildcard'

import { LocalRepository } from '../../../../graphql-operations'
import { callFilePicker, useLocalRepositories, useNewLocalRepositoriesPaths } from '../../../../setup-wizard/components'
import { AppNoItemsState } from '../../components'

import styles from './LocalRepositoriesTab.module.scss'

type Path = string

/**
 * Main entry point for the local repositories setup settings page/tab,
 * provides abilities to specify a local repositories path and see list
 * of resolved repositories by a given path.
 */
export const LocalRepositoriesTab: FC = () => (
    <LocalRepositoriesWidget>
        {api => (
            <PageHeader
                headingElement="h2"
                description="Add your local repositories"
                path={[{ text: 'Local repositories' }]}
                actions={<PathsPickerActions onPathsChange={api.addNewPaths} />}
                className="mb-3"
            />
        )}
    </LocalRepositoriesWidget>
)

interface LocalRepositoriesWidgetProps {
    children: (api: { addNewPaths: (paths: Path[]) => Promise<void> }) => ReactNode
    className?: string
    onRepositoriesChange?: (repositories: LocalRepository[]) => void
}

export const LocalRepositoriesWidget: FC<LocalRepositoriesWidgetProps> = props => {
    const { children, className, onRepositoriesChange = noop } = props

    const {
        paths,
        loading: pathsLoading,
        loaded: pathLoaded,
        error: pathsError,
        addNewPaths,
        deletePath,
    } = useNewLocalRepositoriesPaths()

    const {
        repositories,
        loading: repositoriesLoading,
        loaded: repositoriesLoaded,
        error: repositoriesError,
    } = useLocalRepositories({ paths, skip: paths.length === 0 })

    useLayoutEffect(() => {
        onRepositoriesChange(repositories)
    }, [repositories, onRepositoriesChange])

    const handleRepositoriesDelete = async (pathToDelete: Path): Promise<void> => {
        await deletePath(pathToDelete)
    }

    const anyLoading = pathsLoading || repositoriesLoading
    const anyError = pathsError || repositoriesError
    const allLoaded = pathLoaded && repositoriesLoaded

    return (
        <div className={classNames(className, styles.root)}>
            {children({ addNewPaths })}

            <Container className={styles.container}>
                {anyError && <ErrorAlert error={anyError} />}

                {!anyError && anyLoading && !allLoaded && <LoadingSpinner />}
                {!anyError && allLoaded && repositories.length === 0 && (
                    <AppNoItemsState
                        title="No local repositories"
                        subTitle="Pick local repositories with buttons above"
                    />
                )}
                {!anyError && allLoaded && (
                    <LocalRepositoriesList
                        paths={paths}
                        repositories={repositories}
                        onPathDelete={handleRepositoriesDelete}
                    />
                )}
            </Container>
        </div>
    )
}

interface PathsPickerActionsProps {
    className?: string
    onPathsChange: (paths: string[]) => void
}

/**
 * Local repositories path picker buttons, both do the same job,
 * but we have two buttons to improve user understanding what options
 * they have in the file picker.
 */
export const PathsPickerActions: FC<PathsPickerActionsProps> = ({ className, onPathsChange }) => {
    const handleClickCallPathPicker = async (): Promise<void> => {
        const paths = await callFilePicker()

        if (paths !== null) {
            onPathsChange(paths)
        }
    }

    return (
        <div className={classNames(className, styles.headerActions)}>
            <Button variant="primary" onClick={handleClickCallPathPicker}>
                <Icon svgPath={mdiGit} aria-hidden={true} /> Add a repository
            </Button>
            <Button variant="primary" onClick={handleClickCallPathPicker}>
                <Icon svgPath={mdiFolderMultiplePlusOutline} aria-hidden={true} /> Add all repositories from a folder
            </Button>
        </div>
    )
}

interface Directory {
    path: Path
    repositories: LocalRepository[]
}

interface LocalRepositoriesListProps {
    paths: Path[]
    repositories: LocalRepository[]
    onPathDelete: (path: Path) => void
}

/**
 * Fetches and renders a list of local repositories by a list of specified
 * local paths, it also aggregates list of resolved repositories and group them
 * by directories if repositories are in the same directory.
 *
 * NOTE: at the moment we have to have this group logic on the client since
 * backend API doesn't expose this information, but as soon as we have this in
 * API we can simplify this component.
 */
const LocalRepositoriesList: FC<LocalRepositoriesListProps> = ({ paths, repositories, onPathDelete }) => {
    const { folders, plainRepositories } = useMemo(() => {
        const repositoriesTree: Record<Path, LocalRepository[]> = {}

        for (const path of paths) {
            for (const repository of repositories) {
                // Exactly match in this case path in file picker is a path to
                // .git repository, mark as plain repository
                if (repository.path === path) {
                    repositoriesTree[path] = [repository]
                    continue
                }

                // Repository's path falls within the path, means that path in file picker
                // is a path to directory with multiples .git repositories
                if (repository.path.startsWith(path)) {
                    const existingDirectoryRepositories = repositoriesTree[path] ?? []
                    repositoriesTree[path] = [...existingDirectoryRepositories, repository]
                }
            }
        }

        const folders: Directory[] = []
        const plainRepositories: LocalRepository[] = []

        for (const path of Object.keys(repositoriesTree)) {
            const repositories = repositoriesTree[path]

            if (repositories.length > 1) {
                folders.push({
                    path,
                    repositories,
                })

                continue
            }

            if (repositories.length === 1 && repositories[0].path !== path) {
                folders.push({
                    path,
                    repositories,
                })
            } else {
                plainRepositories.push(repositories[0])
            }
        }

        return { folders, plainRepositories }
    }, [paths, repositories])

    return (
        <ul className={styles.list}>
            {folders.map(folder => (
                <DirectoryItem key={folder.path} directory={folder} onDelete={() => onPathDelete(folder.path)} />
            ))}
            {plainRepositories.map(repository => (
                <RepositoryItem
                    key={repository.path}
                    repository={repository}
                    withDelete={true}
                    onDelete={() => onPathDelete(repository.path)}
                />
            ))}
        </ul>
    )
}

interface DirectoryItemProps {
    directory: Directory
    onDelete: () => void
}

const DirectoryItem: FC<DirectoryItemProps> = ({ directory, onDelete }) => (
    <li className={classNames(styles.listItem, styles.listItemDirectory)}>
        <Collapse>
            {({ isOpen, setOpen }) => (
                <>
                    <div className={styles.listItemDirectoryContentWrapper}>
                        <Button
                            variant="secondary"
                            outline={true}
                            onClick={() => setOpen(!isOpen)}
                            className={styles.listItemDirectoryContent}
                        >
                            <div className={styles.listItemDirectoryPath}>
                                <Icon
                                    aria-hidden={true}
                                    svgPath={mdiFolderMultipleOutline}
                                    inline={false}
                                    className={styles.listItemIcon}
                                />
                                <Text weight="bold" className={styles.listItemName}>
                                    {directory.path}
                                </Text>
                            </div>
                            <Text size="small" className={styles.listItemDescription}>
                                This folder contains {directory.repositories.length}{' '}
                                {pluralize('repository', directory.repositories.length, 'repositories')}
                            </Text>
                        </Button>
                        <Button variant="secondary" onClick={onDelete} className={styles.listItemAction}>
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </div>
                    <CollapsePanel as="ul" className={classNames(styles.list, styles.listSubList)}>
                        {directory.repositories.map(repository => (
                            <RepositoryItem key={repository.path} repository={repository} withDelete={false} />
                        ))}
                    </CollapsePanel>
                </>
            )}
        </Collapse>
    </li>
)

interface RepositoryItemProps {
    repository: LocalRepository
    withDelete: boolean
    onDelete?: () => void
}

const RepositoryItem: FC<RepositoryItemProps> = ({ repository, withDelete, onDelete }) => (
    <li className={styles.listItem}>
        <span className={styles.listItemContent}>
            <Icon aria-hidden={true} svgPath={mdiGit} inline={false} className={styles.listItemIcon} />
            <Text weight="bold" className={styles.listItemName}>
                {repository.name}
            </Text>
            <Text size="small" className={styles.listItemDescription}>
                {repository.path}
            </Text>
        </span>
        {withDelete && (
            <Button variant="secondary" onClick={onDelete} className={styles.listItemAction}>
                <Icon aria-hidden={true} svgPath={mdiDelete} />
            </Button>
        )}
    </li>
)
