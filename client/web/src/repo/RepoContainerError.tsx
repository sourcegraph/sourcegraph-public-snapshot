import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { ErrorLike } from '@sourcegraph/common'
import {
    CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
    isRepoNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { RepoQuestionIcon } from '@sourcegraph/shared/src/components/icons'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { Code, ErrorMessage } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'

import { DirectImportRepoAlert } from './DirectImportRepoAlert'
import { RepositoryNotFoundPage } from './RepositoryNotFoundPage'

interface RepoContainerErrorProps {
    /** The repo fetch error. */
    repoFetchError: ErrorLike

    /** The repository name. */
    repoName: string

    /** Whether the viewer is a site admin. */
    viewerCanAdminister: boolean
}

export const RepoContainerError: React.FunctionComponent<React.PropsWithChildren<RepoContainerErrorProps>> = props => {
    const { repoFetchError, repoName, viewerCanAdminister } = props

    if (isRepoNotFoundErrorLike(repoFetchError)) {
        return <RepositoryNotFoundPage repo={repoName} viewerCanAdminister={viewerCanAdminister} />
    }

    if (isCloneInProgressErrorLike(repoFetchError)) {
        return (
            <HeroPage
                icon={SourceRepositoryIcon}
                title={displayRepoName(repoName)}
                className="repository-cloning-in-progress-page"
                subtitle="Cloning in progress"
                detail={<Code>{(repoFetchError as CloneInProgressError).progress}</Code>}
                body={<DirectImportRepoAlert className="mt-3" />}
            />
        )
    }

    if (isRevisionNotFoundErrorLike(repoFetchError)) {
        return <HeroPage icon={RepoQuestionIcon} title="Empty repository" />
    }

    return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoFetchError} />} />
}
