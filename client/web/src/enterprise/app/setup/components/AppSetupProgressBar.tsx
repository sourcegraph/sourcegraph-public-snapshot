import { FC } from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Text } from '@sourcegraph/wildcard'

import { RepositoriesProgressResult } from '../../../../graphql-operations'

import styles from './AppSetupProgressBar.modules.scss'

const REPO_UPLOADING_PROGRESS = gql`
    query RepositoriesProgress {
        embeddingsSetupProgress {
            overallPercentComplete
            currentRepository
            currentRepositoryFilesProcessed
            currentRepositoryTotalFilesToProcess
        }
    }
`

export const AppSetupProgressBar: FC = props => {
    const { data } = useQuery<RepositoriesProgressResult>(REPO_UPLOADING_PROGRESS, {
        pollInterval: 2000,
        fetchPolicy: 'cache-and-network',
    })

    if (
        !data ||
        data.embeddingsSetupProgress.overallPercentComplete === null ||
        data.embeddingsSetupProgress.overallPercentComplete === 100
    ) {
        return null
    }
    const currentRepository = data.embeddingsSetupProgress.currentRepository
    const filesProcessed = data.embeddingsSetupProgress.currentRepositoryFilesProcessed
    const filesToProcess = data.embeddingsSetupProgress.currentRepositoryTotalFilesToProcess

    const hasDetails = currentRepository && filesProcessed !== null && filesToProcess !== null

    return (
        <div className={styles.root}>
            <div className={styles.description}>
                {hasDetails && (
                    <>
                        <span>Generate embeddings for {data.embeddingsSetupProgress.currentRepository}</span>
                        <Text size="small" className={styles.percent}>
                            {data.embeddingsSetupProgress.currentRepositoryFilesProcessed} /{' '}
                            {data.embeddingsSetupProgress.currentRepositoryTotalFilesToProcess}
                        </Text>
                    </>
                )}

                {!hasDetails && <span>Generate repositories embeddings</span>}
            </div>

            <div className={styles.progress}>
                <div
                    className={styles.bar}
                    style={{ width: `${data.embeddingsSetupProgress.overallPercentComplete}%` }}
                />
            </div>
        </div>
    )
}
