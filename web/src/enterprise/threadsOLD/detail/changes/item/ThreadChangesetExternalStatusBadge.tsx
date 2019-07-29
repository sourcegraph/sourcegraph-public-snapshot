import classNames from 'classnames'
import { upperFirst } from 'lodash'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React from 'react'
import { ChangesetExternalStatus } from '../../backend'

interface Props {
    status: ChangesetExternalStatus
    className?: string
}

/**
 * A badge that displays the external status of a changeset.
 */
export const ThreadChangesetExternalStatusBadge: React.FunctionComponent<Props> = ({ status, className = '' }) => (
    <span
        className={classNames(
            'badge',
            'd-flex',
            'align-items-center',
            {
                'bg-success': status === 'open',
                'bg-purple': status === 'merged',
                'bg-danger': status === 'closed',
            },
            className
        )}
    >
        <SourcePullIcon className="icon-inline mr-1" /> {upperFirst(status)}
    </span>
)
