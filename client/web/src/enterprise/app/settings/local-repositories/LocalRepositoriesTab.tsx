import { FC, ReactNode, useLayoutEffect } from 'react'

import { mdiGit, mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import { Text, Icon, Button, Container, PageHeader, LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import { LocalRepository } from '../../../../graphql-operations'
import { callFilePicker, useLocalRepositories, useNewLocalRepositoriesPaths } from '../../../../setup-wizard/components'
import { NoReposAddedState } from '../../components'

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
                {!anyError && allLoaded && repositories.length === 0 && <NoReposAddedState />}
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
        </div>
    )
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
const LocalRepositoriesList: FC<LocalRepositoriesListProps> = ({ repositories, onPathDelete }) => (
    <ul className={styles.list}>
        {repositories.map(repository => (
            <RepositoryItem
                key={repository.path}
                repository={repository}
                withDelete={true}
                onDelete={() => onPathDelete(repository.path)}
            />
        ))}
    </ul>
)

interface RepositoryItemProps {
    repository: LocalRepository
    withDelete: boolean
    onDelete?: () => void
}

const RepositoryItem: FC<RepositoryItemProps> = ({ repository, withDelete, onDelete }) => (
    <li className={styles.listItem}>
        <span className={styles.listItemContent}>
            <span className={styles.listItemNameWrapper}>
                <Icon aria-hidden={true} svgPath={mdiGit} inline={false} className={styles.listItemIcon} />
                <Text weight="bold" className={styles.listItemName}>
                    {repository.name}
                </Text>
            </span>
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
