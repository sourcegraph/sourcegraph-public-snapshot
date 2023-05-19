import { FC, ReactNode, useMemo } from 'react'

import { mdiFolderMultipleOutline, mdiFolderMultiplePlusOutline, mdiGit, mdiDelete } from '@mdi/js'
import classNames from 'classnames'

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
}

export const LocalRepositoriesWidget: FC<LocalRepositoriesWidgetProps> = props => {
    const { children, className } = props

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
                {!anyError && allLoaded && repositories.length === 0 && <ZeroState />}
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

const ZeroState: FC = () => (
    <div className={styles.zeroState}>
        <svg width="144" height="144" viewBox="0 0 144 144" fill="none" xmlns="http://www.w3.org/2000/svg">
            <g clipPath="url(#clip0_317_910)">
                <path
                    fill="var(--light-part)"
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M112 10C113.309 9.97382 114.555 9.43554 115.471 8.50071C116.388 7.56587 116.901 6.30903 116.901 5C116.901 3.69097 116.388 2.43413 115.471 1.49929C114.555 0.564463 113.309 0.0261766 112 0C110.691 0.0261766 109.445 0.564463 108.529 1.49929C107.612 2.43413 107.099 3.69097 107.099 5C107.099 6.30903 107.612 7.56587 108.529 8.50071C109.445 9.43554 110.691 9.97382 112 10ZM8 5L41 21L24 30L8 5ZM99.5 110H44.5L48.7 84H95.305L99.5 110ZM118 36C118 42.075 113.075 47 107 47C105.375 47.0048 103.77 46.6485 102.3 45.9567C100.83 45.2649 99.5318 44.255 98.5 43L116 29.5C117.26 31.29 118 33.645 118 36Z"
                />
                <path
                    fill="var(--middle-part)"
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M55 45H89L105 144H39L55 45ZM99.5 110H44.5L48.7 84H95.305L99.5 110Z"
                />
                <path
                    fill="var(--dark-part)"
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M71 0H73V5.667L87 15V30H95L89 45H55L49 30H57V15L71 5.667V0ZM64 22C64 17.582 67.584 14 72 14C76.427 14 80 17.582 80 22C80 26.418 76.427 30 72 30C67.584 30 64 26.418 64 22ZM85 58V65H78V58C78 56.5 79.5 55 81.5 55C83.5 55 85 56.5 85 58ZM66 135V144H78V135C78 131.779 75.22 129 72 129C68.78 129 66 131.779 66 135ZM62 101.002V93C62 91.5 60.5 90 58.5 90C56.5 90 55 91.5 55 93V101.002H62Z"
                />
            </g>
            <defs>
                <clipPath id="clip0_317_910">
                    <rect width="144" height="144" fill="white" />
                </clipPath>
            </defs>
        </svg>

        <span className={styles.zeroStateText}>
            <Text className="mb-0">No local repositories</Text>
            <Text size="small" className="mb-0">
                Try pick local repositories with buttons above
            </Text>
        </span>
    </div>
)
