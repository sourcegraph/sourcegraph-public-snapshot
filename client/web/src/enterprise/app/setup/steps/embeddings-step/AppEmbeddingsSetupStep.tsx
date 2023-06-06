import { FC, useContext, ChangeEvent, useState } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Button, H1, ScrollBox, Text, Icon, LoadingSpinner, Tooltip, Link } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../../../../components/FilteredConnection/hooks/useShowMorePagination'
import { SetupRepositoriesListResult } from '../../../../../graphql-operations'
import { EnterprisePageRoutes } from '../../../../../routes.constants'
import { SetupStepsContext, StepComponentProps } from '../../../../../setup-wizard/components'
import { useScheduleRepoEmbeddingJobs } from '../../../../site-admin/cody/backend'
import { AppNoItemsState } from '../../../components'

import styles from './AppEmbeddingsSetupStep.module.scss'

// Due to lack of splitting generate query TypeScript type into separate types
// we have to extract nested types from the generated root query type
type SetupRepository = SetupRepositoriesListResult['repositories']['nodes'][number]

const REPOSITORIES_QUERY = gql`
    query SetupRepositoriesList($first: Int, $after: String) {
        repositories(first: $first, after: $after) {
            __typename
            nodes {
                __typename
                id
                name
                uri
            }
            totalCount
            pageInfo {
                __typename
                hasNextPage
                endCursor
            }
        }
    }
`

export const AppEmbeddingsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    const [scheduleRepoEmbeddingJobs] = useScheduleRepoEmbeddingJobs()
    const [selectedRepository, setSelectedRepository] = useState<SetupRepository | null>(null)

    const { data, connection, loading, hasNextPage, fetchMore } = useShowMorePagination<
        SetupRepositoriesListResult,
        {},
        SetupRepository
    >({
        variables: { first: 50 },
        query: REPOSITORIES_QUERY,
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data) {
                throw new Error('No repositories were found')
            }

            return data.repositories
        },
    })

    const handleNext = () => {
        if (!selectedRepository) {
            return
        }

        scheduleRepoEmbeddingJobs({ variables: { repoNames: [selectedRepository.name] } })
        onNextStep()
    }

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
                    Pick one repository to generate embeddings. This one-off choice supercharges a repo with even better
                    intuition.
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
                {!data && loading && <LoadingSpinner />}

                {connection?.nodes.map(node => (
                    <RepositoryItem
                        key={node.id}
                        checked={selectedRepository?.id === node.id}
                        name={node.name}
                        path={getRepositoryFullPath(node.uri)}
                        onChange={event => setSelectedRepository(node)}
                    />
                ))}

                {hasNextPage && (
                    <Button variant="secondary" outline={true} onClick={() => fetchMore()}>
                        Load more repositories
                        {loading && <LoadingSpinner />}
                    </Button>
                )}

                {connection?.nodes.length === 0 && (
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
            <label className={styles.itemContent}>
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
                    <input type="radio" name="repository" value={name} onChange={onChange} />
                </div>
            </label>
        </li>
    )
}

function getRepositoryFullPath(uri: string): string {
    if (uri.startsWith('/repos')) {
        return uri.slice(6)
    }

    return uri
}
