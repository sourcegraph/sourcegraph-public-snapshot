import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { GroupMembersFields } from '../../../../graphql-operations'
import { formatPersonName } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'

import { GroupDetailContentCardProps } from './GroupDetailContent'

interface Props extends GroupDetailContentCardProps {
    group: GroupMembersFields
    className?: string
}

export const GroupMembersTab: React.FunctionComponent<Props> = ({ group, className }) => (
    <div className={className}>
        {group.members && group.members.length > 0 && (
            <>
                <h4>
                    {group.members.length} {pluralize('member', group.members.length)}
                </h4>
                <ul className="list-unstyled d-flex flex-wrap">
                    {group.members.map(member => (
                        <li key={member.email} className="mr-1 mb-1">
                            <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                                <UserAvatar user={member} size={28} />
                            </LinkOrSpan>
                        </li>
                    ))}
                </ul>
            </>
        )}
    </div>
)
