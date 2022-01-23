import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { Timestamp } from '../../../../../../components/time/Timestamp'
import { GitCommitFields } from '../../../../../../graphql-operations'
import { PersonLink } from '../../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../../user/UserAvatar'

export const LastCommit: React.FunctionComponent<{
    commit: GitCommitFields
    after?: React.ReactFragment
    className?: string
}> = ({ commit, after, className }) => (
    <div className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={commit.author.person} size={18} />
        <PersonLink person={commit.author.person} className="font-weight-bold mr-2 flex-shrink-0" />
        <Link to={commit.url} className="text-truncate flex-grow-1 text-body mr-2" title={commit.message}>
            {commit.subject}
        </Link>
        <small className="text-nowrap text-muted">
            <Link to={commit.url} className="text-monospace text-muted mr-2 d-none d-md-inline">
                {commit.abbreviatedOID}
            </Link>
            <Timestamp date={commit.author.date} noAbout={true} />
        </small>
        {after}
    </div>
)
