import { FC, useContext, ChangeEvent, useState } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

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

import { LocalRepository } from '../../../../../graphql-operations'
import { EnterprisePageRoutes } from '../../../../../routes.constants'
import {
    SetupStepsContext,
    StepComponentProps,
    useLocalRepositories,
    useNewLocalRepositoriesPaths,
} from '../../../../../setup-wizard/components'
import { useScheduleRepoEmbeddingJobs } from '../../../../site-admin/cody/backend'
import { AppNoItemsState } from '../../../components'

import styles from './AppEmbeddingsSetupStep.module.scss'

export const AppEmbeddingsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    const [scheduleRepoEmbeddingJobs] = useScheduleRepoEmbeddingJobs()
    const [selectedRepository, setSelectedRepository] = useState<LocalRepository | null>(null)

    const { paths, loading: pathsLoading, loaded: pathLoaded, error: pathsError } = useNewLocalRepositoriesPaths()

    const {
        repositories,
        loading: repositoriesLoading,
        loaded: repositoriesLoaded,
        error: repositoriesError,
    } = useLocalRepositories({ paths, skip: paths.length === 0 })

    const handleNext = (): void => {
        if (!selectedRepository) {
            return
        }

        scheduleRepoEmbeddingJobs({ variables: { repoNames: [selectedRepository.name] } }).catch(() => {})
        onNextStep()
    }

    const anyLoading = pathsLoading || repositoriesLoading
    const anyError = pathsError || repositoriesError
    const allLoaded = pathLoaded && repositoriesLoaded

    return (
        <div className={classNames(className, styles.root)}>
            <div className={styles.description}>
                <img
                    src="https://storage.googleapis.com/sourcegraph-assets/cody-embeddings.png"
                    alt=""
                    className={styles.descriptionImage}
                />

                <H1 className={styles.descriptionHeading}>Level up your Cody(ing)</H1>

                <Text className={styles.descriptionText}>What’s you preferred repository?</Text>

                <Text className={classNames(styles.descriptionText, styles.descriptionTextSmall)}>
                    Pick one repository to generate embeddings. This supercharges a repo with even better intuition.
                </Text>

                <Text className={classNames(styles.descriptionText, styles.descriptionTextSmall)}>
                    Cody continues to provide quality help across other repositories.
                </Text>

                <Tooltip content={!selectedRepository ? 'Select one repo to continue' : undefined}>
                    <Button
                        size="lg"
                        variant="primary"
                        disabled={!selectedRepository}
                        className={classNames(styles.descriptionNext, {
                            [styles.descriptionNextDisabled]: !selectedRepository,
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
                            checked={selectedRepository?.name === repository.name}
                            name={repository.name}
                            path={repository.path}
                            onChange={event => setSelectedRepository(repository)}
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
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

const RepositoryItem: FC<RepositoryItemProps> = props => {
    const { name, path, checked, onChange } = props

    return (
        <li className={classNames(styles.item, { [styles.itemChecked]: checked })}>
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
                    <input type="radio" name="repository" value={name} onChange={onChange} />
                </div>
            </Label>
        </li>
    )
}
