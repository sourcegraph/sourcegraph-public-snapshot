import { FC, useContext, ChangeEvent, useState, useMemo } from 'react'

import { MutationTuple } from '@apollo/client'
import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { gql, useMutation } from '@sourcegraph/http-client'
import {
    Button,
    H1,
    ScrollBox,
    Text,
    Icon,
    LoadingSpinner,
    Tooltip,
    Link,
    ErrorAlert,
    Label,
} from '@sourcegraph/wildcard'

import {
    ScheduleLocalRepoEmbeddingJobsResult,
    ScheduleLocalRepoEmbeddingJobsVariables,
} from '../../../../../graphql-operations'
import { EnterprisePageRoutes } from '../../../../../routes.constants'
import {
    SetupStepsContext,
    StepComponentProps,
    useLocalRepositories,
    useNewLocalRepositoriesPaths,
} from '../../../../../setup-wizard/components'
import { AppNoItemsState } from '../../../components'

import styles from './AppEmbeddingsSetupStep.module.scss'

type RepoName = string
type EmbeddingsMap = Record<RepoName, boolean>

const SCHEDULE_REPO_EMBEDDING_JOBS = gql`
    mutation ScheduleLocalRepoEmbeddingJobs($repoNames: [String!]!) {
        setupNewAppRepositoriesForEmbedding(repoNames: $repoNames) {
            alwaysNil
        }
    }
`

export function useScheduleRepoEmbeddingJobs(): MutationTuple<
    ScheduleLocalRepoEmbeddingJobsResult,
    ScheduleLocalRepoEmbeddingJobsVariables
> {
    return useMutation<ScheduleLocalRepoEmbeddingJobsResult, ScheduleLocalRepoEmbeddingJobsVariables>(
        SCHEDULE_REPO_EMBEDDING_JOBS
    )
}

export const AppEmbeddingsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    const [scheduleRepoEmbeddingJobs] = useScheduleRepoEmbeddingJobs()
    const [selectedRepositories, setSelectedRepositories] = useState<EmbeddingsMap>({})

    const { paths, loading: pathsLoading, loaded: pathLoaded, error: pathsError } = useNewLocalRepositoriesPaths()

    const {
        repositories,
        loading: repositoriesLoading,
        loaded: repositoriesLoaded,
        error: repositoriesError,
    } = useLocalRepositories({ paths, skip: paths.length === 0 })

    const handleNext = (): void => {
        const selectedRepositoriesNames = Object.entries(selectedRepositories)
            .filter(([name, checked]) => checked)
            .map(([name]) => name)

        if (selectedRepositoriesNames.length === 0) {
            return
        }

        scheduleRepoEmbeddingJobs({ variables: { repoNames: selectedRepositoriesNames } }).catch(() => {})
        onNextStep()
    }

    const handleRepositoriesChecked = (event: ChangeEvent<HTMLInputElement>): void => {
        const repoName = event.target.value
        const checked = event.target.checked

        setSelectedRepositories(embeddings => ({ ...embeddings, [repoName]: checked }))
    }

    const selectedRepositoriesNames = useMemo(
        () =>
            Object.entries(selectedRepositories)
                .filter(([name, checked]) => checked)
                .map(([name]) => name),
        [selectedRepositories]
    )

    const hasAnySelectedRepositories = selectedRepositoriesNames.length > 0
    const hasReachedRepositoriesLimit = selectedRepositoriesNames.length === 10
    const anyLoading = pathsLoading || repositoriesLoading
    const anyError = pathsError || repositoriesError
    const allLoaded = pathLoaded && repositoriesLoaded

    return (
        <div className={classNames(className, styles.root)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>About embeddings</H1>

                <Text className={styles.descriptionText}>
                    Cody uses embeddings, a natural language processing technique, to improve the quality of its
                    responses.
                </Text>

                <Text className={styles.descriptionText}>
                    These embeddings help Cody understand your code better and provide more accurate and relevant
                    assistance
                </Text>

                <Tooltip content={!hasAnySelectedRepositories ? 'Select one repo to continue' : undefined}>
                    <Button
                        size="lg"
                        variant="primary"
                        disabled={!hasAnySelectedRepositories}
                        className={classNames(styles.descriptionNext, {
                            [styles.descriptionNextDisabled]: !hasAnySelectedRepositories,
                        })}
                        onClick={handleNext}
                    >
                        Next →
                    </Button>
                </Tooltip>
            </div>
            <ScrollBox className={styles.repositories} wrapperClassName={styles.repositoriesWrapper}>
                {anyError && <ErrorAlert error={anyError} />}

                {!anyError && anyLoading && !allLoaded && <LoadingSpinner />}
                {!anyError && allLoaded && repositories.length === 0 && (
                    <AppNoItemsState
                        title="No repositories were found"
                        subTitle={
                            <>
                                Try to add local repositories on the{' '}
                                <Link to={`${EnterprisePageRoutes.AppSetup}/local-repositories`}>previous step</Link>
                            </>
                        }
                    />
                )}
                {!anyError &&
                    allLoaded &&
                    repositories.map(repository => (
                        <RepositoryItem
                            key={repository.name}
                            name={repository.name}
                            path={repository.path}
                            checked={selectedRepositories[repository.name]}
                            disabled={hasReachedRepositoriesLimit}
                            onChange={handleRepositoriesChecked}
                        />
                    ))}
            </ScrollBox>
        </div>
    )
}

interface RepositoryItemProps {
    name: string
    path: string
    checked: boolean
    disabled: boolean
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

const RepositoryItem: FC<RepositoryItemProps> = props => {
    const { name, path, checked, disabled, onChange } = props

    return (
        <Tooltip
            placement="topEnd"
            content={disabled && !checked ? "You've reached the maximum of 10 repositories" : null}
        >
            <li
                className={classNames(styles.item, {
                    [styles.itemChecked]: checked,
                    [styles.itemDisabled]: disabled && !checked,
                })}
            >
                <Label className={styles.itemContent}>
                    <div className={styles.itemDescription}>
                        <Icon inline={false} svgPath={mdiGit} aria-hidden={true} className={styles.itemIcon} />
                        <span className={styles.itemDescrtiptionText}>
                            <Text className={styles.itemText}>{name}</Text>
                            <Text size="small" className={classNames('text-muted', styles.itemText)}>
                                {path}
                            </Text>
                        </span>
                    </div>
                    <div className={styles.itemControl}>
                        {checked && (
                            <Text size="small" className="m-0">
                                Improved context
                            </Text>
                        )}
                        {/* eslint-disable-next-line react/forbid-elements */}
                        <input type="checkbox" name="repository" value={name} checked={checked} onChange={onChange} />
                    </div>
                </Label>
            </li>
        </Tooltip>
    )
}
