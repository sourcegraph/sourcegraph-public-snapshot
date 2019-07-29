import H from 'history'
import FileFindIcon from 'mdi-react/FileFindIcon'
import MessageOutlineIcon from 'mdi-react/MessageOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../shared/src/util/strings'
import { Timestamp } from '../../../components/time/Timestamp'
import { PersonLink } from '../../../user/PersonLink'
import { ThreadStatusIcon } from '../components/threadStatus/ThreadStatusIcon'
import { ThreadsListContext } from './ThreadsList'

interface Props extends ThreadsListContext {
    thread: GQL.IDiscussionThread
    location: H.Location
}

/**
 * A list item for a thread in {@link ThreadsList}.
 */
export const ThreadsListItem: React.FunctionComponent<Props> = ({ thread, itemCheckboxes, location }) => (
    <li className="list-group-item p-2">
        <div className="d-flex align-items-start pl-1">
            {itemCheckboxes && (
                <div
                    className="form-check ml-1 mr-2"
                    /* tslint:disable-next-line:jsx-ban-props */
                    style={{ marginTop: '2px' /* stylelint-disable-line declaration-property-unit-whitelist */ }}
                >
                    <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
                </div>
            )}
            <ThreadStatusIcon thread={thread} className="small mr-2 mt-2" />
            <div className="flex-1">
                <div className="d-flex align-items-center flex-wrap">
                    <h3 className="d-flex align-items-center mb-0 mr-2">
                        <Link to={thread.url} className="text-body">
                            {thread.title}
                        </Link>
                        <span className="badge badge-secondary ml-1 d-none">123</span> {/* TODO!(sqs) */}
                    </h3>
                    <div>
                        {thread.title
                            .toLowerCase()
                            .split(' ')
                            .filter(w => w.length >= 3)
                            .map((label, i) => (
                                <span key={i} className={`badge mr-1 ${badgeColorClass(label)}`}>
                                    {label}
                                </span>
                            ))}
                    </div>
                </div>
                <ul className="list-unstyled d-flex align-items-center small text-muted mb-0">
                    <li>
                        #{thread.idWithoutKind} created <Timestamp date={thread.createdAt} /> by{' '}
                        <PersonLink user={thread.author} />
                    </li>
                    {thread.targets.totalCount > 0 && (
                        <li className="ml-2 d-flex align-items-center">
                            <FileFindIcon className="icon-inline mr-1" /> {thread.targets.totalCount}{' '}
                            {pluralize('item', thread.targets.totalCount)}
                        </li>
                    )}
                </ul>
            </div>
            <div>
                <ul className="list-inline d-flex align-items-center">
                    {thread.comments.totalCount > 0 && (
                        <li className="list-inline-item">
                            <small className="text-muted">
                                <MessageOutlineIcon className="icon-inline" /> {thread.comments.totalCount}
                            </small>
                        </li>
                    )}
                </ul>
            </div>
        </div>
    </li>
)

function badgeColorClass(label: string): string {
    if (label === 'security' || label.endsWith('sec')) {
        return 'badge-danger'
    }
    const CLASSES = ['badge-primary', 'badge-warning', 'badge-info', 'badge-success', 'badge-danger']
    const k = label.split('').reduce((sum, c) => (sum += c.charCodeAt(0)), 0)
    return CLASSES[k % CLASSES.length]
}
