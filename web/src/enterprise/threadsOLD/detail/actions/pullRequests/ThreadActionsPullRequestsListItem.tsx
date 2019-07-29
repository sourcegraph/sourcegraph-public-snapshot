import classNames from 'classnames'
import formatDistance from 'date-fns/formatDistance'
import { upperFirst } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseCircleIcon from 'mdi-react/CloseCircleIcon'
import DotsHorizontalCircleIcon from 'mdi-react/DotsHorizontalCircleIcon'
import FileFindIcon from 'mdi-react/FileFindIcon'
import MessageOutlineIcon from 'mdi-react/MessageOutlineIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../../../shared/src/util/strings'
import { PullRequest, ThreadSettings } from '../../../settings'
import { CreatePRButton } from './CreatePRButton'

interface Props extends ExtensionsControllerProps {
    pull: PullRequest
    thread: Pick<GQL.IDiscussionThread, 'id' | 'url'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
}

/**
 * A row for single GitHub pull request in a thread.
 */
export const ThreadActionsPullRequestsListItem: React.FunctionComponent<Props> = ({
    pull,
    thread,
    onThreadUpdate,
    threadSettings,
    className = '',
    extensionsController,
}) => (
    <div className={`${className}`}>
        <div className="d-flex align-items-start">
            <div
                className="form-check mx-2"
                /* tslint:disable-next-line:jsx-ban-props */
                style={{ marginTop: '2px' /* stylelint-disable-line declaration-property-unit-whitelist */ }}
            >
                <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
            </div>
            <SourcePullIcon
                className={classNames('icon-inline', 'mr-2', 'h5', 'mb-0', {
                    'text-success': pull.status === 'open',
                    'text-purple': pull.status === 'merged',
                    'text-danger': pull.status === 'closed',
                    'text-warning': pull.status === 'pending',
                })}
                data-tooltip={upperFirst(pull.status)}
            />
            <div className="flex-1">
                <h3 className="d-flex align-items-center mb-0">
                    <LinkOrSpan
                        to={pull.number === undefined ? undefined : `https://${pull.repo}/pull/${pull.number}`}
                        target="_blank"
                        // tslint:disable-next-line:jsx-ban-props
                        style={{ color: 'var(--body-color)' }}
                    >
                        {pull.label ? (
                            <>
                                <span className="font-weight-normal">{displayRepoName(pull.repo)}</span> &mdash;{' '}
                                <code>{pull.label}</code>
                            </>
                        ) : (
                            displayRepoName(pull.repo)
                        )}
                    </LinkOrSpan>
                </h3>
                <ul className="list-unstyled d-flex align-items-center small text-muted mb-0">
                    {pull.status !== 'pending' && (
                        <li>
                            <small className="text-muted">
                                #{pull.number} updated {formatDistance(Date.parse(pull.updatedAt), Date.now())} ago by{' '}
                                <strong>{pull.updatedBy}</strong>
                            </small>
                        </li>
                    )}
                    <li className="ml-2 d-flex align-items-center">
                        <Link
                            to={`${thread.url}/inbox?q=${encodeURIComponent(`repo:${pull.repo}`)}`}
                            className="text-decoration-none"
                        >
                            <FileFindIcon className="icon-inline" /> {pull.items.length}{' '}
                            {pluralize('change', pull.items.length)}
                        </Link>
                    </li>
                </ul>
            </div>
            <div>
                {pull.status === 'pending' ? (
                    <CreatePRButton
                        pull={pull}
                        thread={thread}
                        onThreadUpdate={onThreadUpdate}
                        threadSettings={threadSettings}
                        extensionsController={extensionsController}
                    />
                ) : (
                    pull.commentsCount > 0 && (
                        <ul className="list-inline d-flex align-items-center">
                            <li className="list-inline-item">
                                <small className="text-muted">
                                    <MessageOutlineIcon className="icon-inline" /> {pull.commentsCount}
                                </small>
                            </li>
                        </ul>
                    )
                )}
            </div>
        </div>
    </div>
)
