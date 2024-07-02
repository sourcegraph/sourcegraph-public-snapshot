import { useEffect } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import type { ErrorLike } from '@sourcegraph/common'
import {
    type CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
    isRepoNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { RepoQuestionIcon } from '@sourcegraph/shared/src/components/icons'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Code, ErrorMessage, Link, Text } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'

import { DirectImportRepoAlert } from './DirectImportRepoAlert'
import { RepositoryNotFoundPage } from './RepositoryNotFoundPage'

interface RepoContainerErrorProps extends TelemetryV2Props {
    /** The repo fetch error. */
    repoFetchError: ErrorLike

    /** The repository name. */
    repoName: string

    /** Whether the viewer is a site admin. */
    viewerCanAdminister: boolean
}

export const RepoContainerError: React.FunctionComponent<React.PropsWithChildren<RepoContainerErrorProps>> = props => {
    const { repoFetchError, repoName, viewerCanAdminister, telemetryRecorder } = props

    if (isRepoNotFoundErrorLike(repoFetchError)) {
        return (
            <RepositoryNotFoundPage
                repo={repoName}
                viewerCanAdminister={viewerCanAdminister}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    if (isCloneInProgressErrorLike(repoFetchError)) {
        return (
            <CloneInProgressPage
                repoName={repoName}
                viewerCanAdminister={viewerCanAdminister}
                repoFetchError={repoFetchError}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    if (isRevisionNotFoundErrorLike(repoFetchError)) {
        return (
            <RevisionNotFoundErrorPage
                repoName={repoName}
                viewerCanAdminister={viewerCanAdminister}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    return <OtherRepoErrorPage repoFetchError={repoFetchError} telemetryRecorder={telemetryRecorder} />
}

export const CloneInProgressPage: React.FunctionComponent<React.PropsWithChildren<RepoContainerErrorProps>> = props => {
    const { repoName, viewerCanAdminister, repoFetchError, telemetryRecorder } = props

    useEffect(() => telemetryRecorder.recordEvent('repo.error.cloneInProgress', 'view'), [telemetryRecorder])
    return (
        <HeroPage
            icon={SourceRepositoryIcon}
            title={displayRepoName(repoName)}
            className="repository-cloning-in-progress-page"
            subtitle={<Text>Cloning in progress</Text>}
            detail={
                <>
                    <Code>{(repoFetchError as CloneInProgressError).progress}</Code>
                    {viewerCanAdminister && (
                        <Text className="mt-4">
                            <Link to={`${repoName}/-/settings`}>Go to settings</Link> to view details
                        </Text>
                    )}
                </>
            }
            body={<DirectImportRepoAlert className="mt-3" />}
        />
    )
}

export const RevisionNotFoundErrorPage: React.FunctionComponent<
    React.PropsWithChildren<Pick<RepoContainerErrorProps, 'repoName' | 'viewerCanAdminister' | 'telemetryRecorder'>>
> = props => {
    const { repoName, viewerCanAdminister, telemetryRecorder } = props

    useEffect(() => telemetryRecorder.recordEvent('repo.error.revisionNotFound', 'view'), [telemetryRecorder])
    return (
        <HeroPage
            icon={RepoQuestionIcon}
            title="Empty repository"
            detail={
                <>
                    {viewerCanAdminister && (
                        <Text>
                            <Link to={`${repoName}/-/settings`}>Go to settings</Link>
                        </Text>
                    )}
                </>
            }
        />
    )
}

export const OtherRepoErrorPage: React.FunctionComponent<
    React.PropsWithChildren<Pick<RepoContainerErrorProps, 'repoFetchError' | 'telemetryRecorder'>>
> = props => {
    const { repoFetchError, telemetryRecorder } = props

    useEffect(() => telemetryRecorder.recordEvent('repo.error.other', 'view'), [telemetryRecorder])
    return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoFetchError} />} />
}
