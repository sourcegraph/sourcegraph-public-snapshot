import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Page } from '../../../../components/Page'
import { GroupMembersFields } from '../../../../graphql-operations'
import { formatPersonName } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'

import { GroupDetailContentCardProps } from './GroupDetailContent'

interface Props extends GroupDetailContentCardProps {
    group: GroupMembersFields
    className?: string
}

export const GroupMembersTab: React.FunctionComponent<Props> = ({ group, className }) => (
    <Page className={className}>
        {group.members && group.members.length > 0 && (
            <>
                <h4>
                    {group.members.length} {pluralize('member', group.members.length)}
                </h4>
                <ul className="list-group">
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
    </Page>
)
