import type { FC } from 'react'

import classNames from 'classnames'

import { gql, useMutation } from '@sourcegraph/http-client'
import { Button, type ButtonProps, LoadingSpinner } from '@sourcegraph/wildcard'

import { CodeInsightsJobFragment } from '../../query'

import styles from './CodeInsightsJobsActions.module.scss'

interface CodeInsightsJobActionsProps {
    selectedJobIds: string[]
    className?: string
    onSelectionClear: () => void
}

export const CodeInsightsJobsActions: FC<CodeInsightsJobActionsProps> = props => {
    const { selectedJobIds, className, onSelectionClear } = props

    const [retryJobs, { loading: retryLoading }] = useMutation(getMultipleRetryMutation(selectedJobIds), {
        refetchQueries: ['GetCodeInsightsJobs'],
        onCompleted: onSelectionClear,
    })
    const [moveToBack, { loading: moveToBackLoading }] = useMutation(getMultipleMovetoBackMutation(selectedJobIds), {
        refetchQueries: ['GetCodeInsightsJobs'],
        onCompleted: onSelectionClear,
    })
    const [moveToFront, { loading: moveToFrontLoading }] = useMutation(getMultipleMovetoFrontMutation(selectedJobIds), {
        refetchQueries: ['GetCodeInsightsJobs'],
        onCompleted: onSelectionClear,
    })

    const loading = retryLoading || moveToBackLoading || moveToFrontLoading

    return (
        <div className={classNames(className, styles.actions)}>
            <JobActionButton
                disabled={loading}
                loading={retryLoading}
                actionCount={selectedJobIds.length}
                onClick={() => retryJobs()}
            >
                Retry
            </JobActionButton>
            <JobActionButton
                disabled={loading}
                loading={moveToFrontLoading}
                actionCount={selectedJobIds.length}
                onClick={() => moveToBack()}
            >
                Back of queue
            </JobActionButton>
            <JobActionButton
                disabled={loading}
                loading={moveToFrontLoading}
                actionCount={selectedJobIds.length}
                onClick={() => moveToFront()}
            >
                Front of queue
            </JobActionButton>
            {selectedJobIds.length > 0 && (
                <Button variant="secondary" outline={true} onClick={onSelectionClear}>
                    Clear selection
                </Button>
            )}
        </div>
    )
}

interface JobActionButton extends ButtonProps {
    actionCount: number
    loading: boolean
    disabled: boolean
}

const JobActionButton: FC<JobActionButton> = props => {
    const { actionCount, children, loading } = props

    return (
        <Button {...props} variant="primary" disabled={actionCount === 0 || loading} className={styles.action}>
            {actionCount > 0 && <span className={styles.count}>{actionCount}</span>}
            {children}
            {loading && <LoadingSpinner className="ml-2" />}
        </Button>
    )
}

const BLANK_MUTATION = gql`
    mutation RetryInsightSeriesBackfillBlank {
        retryInsightSeriesBackfill(id: "0001") {
            id
        }
    }
`

function getMultipleRetryMutation(jobIds: string[]): string {
    if (jobIds.length === 0) {
        return BLANK_MUTATION
    }

    const mutations = jobIds
        .map(
            (id, index) => `

        retry${index}: retryInsightSeriesBackfill(id: "${id}") {
           ...InsightJob
        }

    `
        )
        .join(' ')

    return gql`
        ${CodeInsightsJobFragment}
        mutation RetryCodeInsightsJobs {
            __typename
            ${mutations}
        }
    `
}

function getMultipleMovetoFrontMutation(jobIds: string[]): string {
    if (jobIds.length === 0) {
        return BLANK_MUTATION
    }

    const mutations = jobIds
        .map(
            (id, index) => `

        moveJobToFront${index}: moveInsightSeriesBackfillToFrontOfQueue(id: "${id}") {
           ...InsightJob
        }

    `
        )
        .join(' ')

    return gql`
        ${CodeInsightsJobFragment}
        mutation MoveToFrontCodeInsightsJobs {
            __typename
            ${mutations}
        }
    `
}

function getMultipleMovetoBackMutation(jobIds: string[]): string {
    if (jobIds.length === 0) {
        return BLANK_MUTATION
    }

    const mutations = jobIds
        .map(
            (id, index) => `

        moveJobToBack${index}: moveInsightSeriesBackfillToBackOfQueue(id: "${id}") {
           ...InsightJob
        }

    `
        )
        .join(' ')

    return gql`
        ${CodeInsightsJobFragment}
        mutation MoveToBackCodeInsighsJobs {
            __typename
            ${mutations}
        }
    `
}
