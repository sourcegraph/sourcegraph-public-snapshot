import React, { useMemo, useState } from 'react'
import { DropdownItem, DropdownMenu, DropdownMenuProps } from 'reactstrap'
import { displayRepoName } from '../../../../../../../../../shared/src/components/RepoFileLink'
import * as GQL from '../../../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../../../shared/src/util/errors'
import { PullRequest, ThreadSettings } from '../../../../../settings'

interface Props extends Pick<DropdownMenuProps, 'right'> {
    threadSettings: ThreadSettings
    inboxItem: GQL.IDiscussionThreadTargetRepo

    /** Called when the user clicks to create a new pull request to add to. */
    onCreateClick: () => void

    /** Called when the user clicks on an existing pull request to add to. */
    onAddToExistingClick: (pull: PullRequest) => void
}

const LOADING: 'loading' = 'loading'

const queryPullRequests = async (
    threadSettings: ThreadSettings,
    inboxItem: GQL.IDiscussionThreadTargetRepo
): Promise<PullRequest[]> => (threadSettings.pullRequests || []).filter(pull => pull.repo === inboxItem.repository.name)

/**
 * A dropdown menu with a list of pull requests and an option to create a new pull request.
 */
export const PullRequestDropdownMenu: React.FunctionComponent<Props> = ({
    threadSettings,
    inboxItem,
    onCreateClick,
    onAddToExistingClick,
    ...props
}) => {
    const [pullRequestsOrError, setPullRequestsOrError] = useState<typeof LOADING | PullRequest[] | ErrorLike>(LOADING)

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setPullRequestsOrError(await queryPullRequests(threadSettings, inboxItem))
        } catch (err) {
            setPullRequestsOrError(asError(err))
        }
    }, [threadSettings, inboxItem])

    return (
        <DropdownMenu {...props}>
            {pullRequestsOrError === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading pull requests...
                </DropdownItem>
            ) : isErrorLike(pullRequestsOrError) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading existing pull requests
                </DropdownItem>
            ) : (
                pullRequestsOrError.length > 0 && (
                    <>
                        <DropdownItem header={true} className="py-1">
                            Add to existing pull request...
                        </DropdownItem>
                        {pullRequestsOrError.map((pull, i) => (
                            <DropdownItem
                                key={i}
                                // tslint:disable-next-line: jsx-no-lambda
                                onClick={() => onAddToExistingClick(pull)}
                                className="d-flex justify-content-between"
                            >
                                {displayRepoName(pull.repo)}{' '}
                                <span className="text-muted ml-3">
                                    {pull.number !== undefined ? `#${pull.number}` : 'pending'}
                                </span>
                            </DropdownItem>
                        ))}
                    </>
                )
            )}
            {pullRequestsOrError === LOADING || isErrorLike(pullRequestsOrError) || pullRequestsOrError.length > 0 ? (
                <DropdownItem divider={true} />
            ) : null}
            <DropdownItem className="dropdown-item" onClick={onCreateClick}>
                New pull request
            </DropdownItem>
        </DropdownMenu>
    )
}
