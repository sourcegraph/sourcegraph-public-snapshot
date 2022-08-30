import { FunctionComponent, PropsWithChildren } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import {
    CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { RepoQuestionIcon } from '@sourcegraph/shared/src/components/icons'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { Code } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'

import { ResolvedRevision } from './backend'
import { DirectImportRepoAlert } from './DirectImportRepoAlert'

export const RepositoryCloningInProgressPage: React.FunctionComponent<{ repoName: string; progress?: string }> = ({
    repoName,
    progress,
}) => (
    <HeroPage
        icon={SourceRepositoryIcon}
        title={displayRepoName(repoName)}
        className="repository-cloning-in-progress-page"
        subtitle="Cloning in progress"
        detail={<Code>{progress}</Code>}
        body={<DirectImportRepoAlert className="mt-3" />}
    />
)

export const EmptyRepositoryPage: FunctionComponent<PropsWithChildren<unknown>> = () => (
    <HeroPage icon={RepoQuestionIcon} title="Empty repository" />
)

interface RepositoryGitDataContainerProps {
    /** The repository. */
    resolvedRevisionOrError: ResolvedRevision | ErrorLike | undefined

    /** The repository name. */
    repoName: string

    /** The fragment to render if the repository's Git data is accessible. */
    children: React.ReactNode
}

/**
 * A container for a repository page that incorporates global Git data but is not tied to one specific revision. A
 * loading/error page is shown if the repository is not yet cloned or is empty. Otherwise, the children are
 * rendered.
 */
export const RepositoryGitDataContainer: React.FunctionComponent<
    React.PropsWithChildren<RepositoryGitDataContainerProps>
> = props => {
    const { resolvedRevisionOrError, repoName, children } = props

    if (isErrorLike(resolvedRevisionOrError)) {
        // Show error page
        if (isCloneInProgressErrorLike(resolvedRevisionOrError)) {
            return (
                <RepositoryCloningInProgressPage
                    repoName={repoName}
                    progress={(resolvedRevisionOrError as CloneInProgressError).progress}
                />
            )
        }

        if (isRevisionNotFoundErrorLike(resolvedRevisionOrError)) {
            return <EmptyRepositoryPage />
        }

        return (
            <HeroPage
                icon={AlertCircleIcon}
                title="Error"
                subtitle={<ErrorMessage error={resolvedRevisionOrError} />}
            />
        )
    }

    return <>{children}</>
}
