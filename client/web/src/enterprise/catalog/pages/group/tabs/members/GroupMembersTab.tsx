import classNames from 'classnames'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { GroupMembersFields } from '../../../../../../graphql-operations'
import { formatPersonName } from '../../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../../user/UserAvatar'

import { GroupDetailContentCardProps } from '../../GroupContent'

interface Props extends GroupDetailContentCardProps {
    group: GroupMembersFields
    className?: string
}

export const GroupMembersTab: React.FunctionComponent<Props> = ({ group, className }) => (
    <div className={classNames('container p-3', className)}>
        {group.members && group.members.length > 0 && (
            <>
                <h4>
                    {group.members.length} {pluralize('member', group.members.length)}
                </h4>
                <ul className="list-group card">
                    {group.members.map(member => (
                        <li key={member.email} className="list-group-item">
                            <LinkOrSpan to={member.user?.url}>
                                <UserAvatar user={member} size={28} className="mr-2" /> {formatPersonName(member)}
                            </LinkOrSpan>
                        </li>
                    ))}
                </ul>
            </>
        )}
    </div>
)
