import { type FC, type ReactNode, useLayoutEffect, useMemo } from 'react'

import { mdiGit, mdiDelete, mdiFolderMultipleOutline } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import { pluralize } from '@sourcegraph/common'
import {
    Text,
    Icon,
    Button,
    Container,
    PageHeader,
    LoadingSpinner,
    ErrorAlert,
    Collapse,
    CollapsePanel,
} from '@sourcegraph/wildcard'

import type { LocalRepository } from '../../../../graphql-operations'
import { callFilePicker, useLocalExternalServices, type LocalCodeHost } from '../../../../setup-wizard/components'
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

    const { services, loading, loaded, error, addRepositories, deleteService } = useLocalExternalServices()

    useLayoutEffect(() => {
        onRepositoriesChange(services.flatMap(service => service.repositories))
    }, [services, onRepositoriesChange])

    return (
        <div className={classNames(className, styles.root)}>
            {children({ addNewPaths: addRepositories })}

            <Container className={styles.container}>
                {error && <ErrorAlert error={error} />}

                {!error && loading && !loaded && <LoadingSpinner />}
                {!error && loaded && services.length === 0 && <NoReposAddedState />}
                {!error && loaded && <LocalRepositoriesList services={services} onPathDelete={deleteService} />}
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
    services: LocalCodeHost[]
    onPathDelete: (service: LocalCodeHost) => void
}

/**
 * Fetches and renders a list of local repositories by a list of specified
 * local paths, it also aggregates list of resolved repositories and group them
 * by directories if repositories are in the same directory.
 */
const LocalRepositoriesList: FC<LocalRepositoriesListProps> = ({ services, onPathDelete }) => {
    const sortedServices = useMemo(
        () =>
            services.slice().sort((serviceA, serviceB) => {
                const result = +serviceB.isFolder - +serviceA.isFolder
                return result === 0 ? serviceA.path.localeCompare(serviceB.path) : result
            }),
        [services]
    )

    return (
        <ul className={styles.list}>
            {sortedServices.map(service =>
                service.isFolder ? (
                    <DirectoryItem key={service.id} service={service} onDelete={onPathDelete} />
                ) : (
                    <RepositoryItem
                        key={service.id}
                        service={service}
                        repository={service.repositories[0]}
                        onDelete={onPathDelete}
                    />
                )
            )}
        </ul>
    )
}

interface DirectoryItemProps {
    service: LocalCodeHost
    onDelete: (service: LocalCodeHost) => void
}

const DirectoryItem: FC<DirectoryItemProps> = ({ service, onDelete }) => (
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
                                    {service.path}
                                </Text>
                            </div>
                            <Text size="small" className={styles.listItemDescription}>
                                This folder contains {service.repositories.length}{' '}
                                {pluralize('repository', service.repositories.length, 'repositories')}
                            </Text>
                        </Button>
                        {!service.autogenerated && (
                            <Button
                                variant="secondary"
                                onClick={() => onDelete(service)}
                                className={styles.listItemAction}
                            >
                                <Icon aria-hidden={true} svgPath={mdiDelete} />
                            </Button>
                        )}
                    </div>
                    <CollapsePanel as="ul" className={classNames(styles.list, styles.listSubList)}>
                        {service.repositories.map(repository => (
                            <RepositoryItem key={repository.path} service={service} repository={repository} />
                        ))}
                    </CollapsePanel>
                </>
            )}
        </Collapse>
    </li>
)

interface RepositoryItemProps {
    service: LocalCodeHost
    repository: LocalRepository
    onDelete?: (service: LocalCodeHost) => void
}

const RepositoryItem: FC<RepositoryItemProps> = ({ service, repository, onDelete }) => (
    <li className={styles.listItem}>
        <span className={styles.listItemContent}>
            <div className={styles.listItemDirectoryPath}>
                <Icon aria-hidden={true} svgPath={mdiGit} inline={false} className={styles.listItemIcon} />
                <Text weight="bold" className={styles.listItemName}>
                    {repository.name}
                </Text>
            </div>

            <Text size="small" className={styles.listItemDescription}>
                {repository.path}
            </Text>
        </span>
        {onDelete && !service.autogenerated && (
            <Button variant="secondary" onClick={() => onDelete(service)} className={styles.listItemAction}>
                <Icon aria-hidden={true} svgPath={mdiDelete} />
            </Button>
        )}
    </li>
)
