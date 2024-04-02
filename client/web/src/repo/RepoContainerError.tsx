import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import AlertIcon from 'mdi-react/AlertIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import type { ErrorLike } from '@sourcegraph/common'
import {
    type CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
    isRepoNotFoundErrorLike,
    isRepoDeniedErrorLike,
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

    if (isRepoDeniedErrorLike(repoFetchError)) {
        telemetryRecorder.recordEvent('repo.error.repoDenied', 'view')
        return (
            <HeroPage
                icon={AlertIcon}
                title={displayRepoName(repoName)}
                body={<Text className="mt-4">Repository cannot be added on-demand: {repoFetchError.reason}.</Text>}
            />
        )
    }

    if (isCloneInProgressErrorLike(repoFetchError)) {
        telemetryRecorder.recordEvent('repo.error.cloneInProgress', 'view')
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

    if (isRevisionNotFoundErrorLike(repoFetchError)) {
        telemetryRecorder.recordEvent('repo.error.revisionNotFound', 'view')
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

    telemetryRecorder.recordEvent('repo.error.other', 'view')
    return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoFetchError} />} />
}
