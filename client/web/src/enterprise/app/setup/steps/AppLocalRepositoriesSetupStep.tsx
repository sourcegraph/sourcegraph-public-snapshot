import { type FC, useContext, useState } from 'react'

import type { MutationTuple } from '@apollo/client'
import { mdiGit } from '@mdi/js'
import classNames from 'classnames'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { gql, useMutation } from '@sourcegraph/http-client'
import { Button, H1, Icon, Text, Tooltip, Link, LoadingSpinner } from '@sourcegraph/wildcard'

import type {
    LocalRepository,
    ScheduleLocalRepoEmbeddingJobsResult,
    ScheduleLocalRepoEmbeddingJobsVariables,
} from '../../../../graphql-operations'
import { callFilePicker, SetupStepsContext, type StepComponentProps } from '../../../../setup-wizard/components'
import { LocalRepositoriesWidget } from '../../settings/local-repositories/LocalRepositoriesTab'

import styles from './AppLocalRepositoriesSetupStep.module.scss'

const MAX_NUMBER_OF_REPOSITORIES = 10

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

export const AddLocalRepositoriesSetupPage: FC<StepComponentProps> = ({ className, setStepId }) => {
    const { onNextStep } = useContext(SetupStepsContext)
    const [scheduleEmbeddings, { loading }] = useScheduleRepoEmbeddingJobs()
    const [repositories, setRepositories] = useState<LocalRepository[]>([])
    const [searchParams] = useSearchParams()
    const navigate = useNavigate()

    const handleNext = (): void => {
        scheduleEmbeddings({
            variables: { repoNames: repositories.map(repo => repo.name) },
        }).catch(() => {})

        // skip install-extensions step if we are coming from vscode
        if (searchParams.get('from') === 'vscode') {
            setStepId?.('all-set')
            navigate('/app-setup/all-set?from=vscode')
        } else {
            onNextStep()
        }
    }

    return (
        <div className={classNames(className, styles.root)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Build your code graph</H1>

                <Text className={styles.descriptionText}>
                    Select up to {MAX_NUMBER_OF_REPOSITORIES} repositories to add to your code graph.
                </Text>

                <Text className={styles.descriptionText}>
                    This code will be sent to OpenAI to create{' '}
                    <Link
                        to="https://docs.sourcegraph.com/cody/explanations/code_graph_context#embeddings"
                        target="_blank"
                        rel="noopener"
                    >
                        embeddings
                    </Link>
                    , which helps Cody build the code graph and generate more accurate answers about your code.
                </Text>

                <Tooltip
                    content={
                        repositories.length > MAX_NUMBER_OF_REPOSITORIES
                            ? `Select fewer repositories, Currently Cody supports a maximum of ${MAX_NUMBER_OF_REPOSITORIES} repositories`
                            : repositories.length === 0
                            ? 'Select at least one repository to continue'
                            : undefined
                    }
                >
                    <Button
                        size="lg"
                        variant="primary"
                        disabled={
                            loading || repositories.length === 0 || repositories.length > MAX_NUMBER_OF_REPOSITORIES
                        }
                        className={styles.descriptionNext}
                        onClick={handleNext}
                    >
                        {loading && <LoadingSpinner />} Next →
                    </Button>
                </Tooltip>
            </div>
            <div className={styles.localRepositories}>
                <LocalRepositoriesWidget
                    className={styles.localRepositoriesWidget}
                    onRepositoriesChange={setRepositories}
                >
                    {api => (
                        <PathsPickerActions
                            disabled={repositories.length >= MAX_NUMBER_OF_REPOSITORIES}
                            className={styles.localRepositoriesButtonsGroup}
                            onPathsChange={api.addNewPaths}
                        />
                    )}
                </LocalRepositoriesWidget>
            </div>
        </div>
    )
}

interface PathsPickerActionsProps {
    disabled: boolean
    className?: string
    onPathsChange: (paths: string[]) => void
}

/**
 * Local repositories path picker buttons, both do the same job,
 * but we have two buttons to improve user understanding what options
 * they have in the file picker.
 */
const PathsPickerActions: FC<PathsPickerActionsProps> = props => {
    const { disabled, className, onPathsChange } = props

    const handleClickCallPathPicker = async (): Promise<void> => {
        const paths = await callFilePicker({ multiple: false })

        if (paths !== null) {
            onPathsChange(paths)
        }
    }

    return (
        <Tooltip
            content={disabled ? `Currently Cody supports a maximum of ${MAX_NUMBER_OF_REPOSITORIES} repositories` : ''}
        >
            <Button variant="primary" disabled={disabled} className={className} onClick={handleClickCallPathPicker}>
                <Icon svgPath={mdiGit} aria-hidden={true} /> Add a repository
            </Button>
        </Tooltip>
    )
}
