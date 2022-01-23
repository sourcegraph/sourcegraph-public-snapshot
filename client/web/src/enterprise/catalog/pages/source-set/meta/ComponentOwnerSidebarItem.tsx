import React from 'react'

import { isDefined } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { ComponentOwnerFields } from '../../../../../graphql-operations'
import { formatPersonName, personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import {
    ComponentOwnerLink,
    COMPONENT_OWNER_LINK_FRAGMENT,
} from '../../../components/component-owner-link/ComponentOwnerLink'
import { GROUP_LINK_FRAGMENT } from '../../../components/group-link/GroupLink'

import styles from './ComponentOwnerSidebarItem.module.scss'

export const COMPONENT_OWNER_FRAGMENT = gql`
    fragment ComponentOwnerFields on Component {
        owner {
            ...ComponentOwnerLinkFields
            ... on Group {
                id
                members {
                    ...PersonLinkFields
                    avatarURL
                }
                ancestorGroups {
                    __typename
                    id
                    ...GroupLinkFields
                }
            }
        }
    }
    ${COMPONENT_OWNER_LINK_FRAGMENT}
    ${personLinkFieldsFragment}
    ${GROUP_LINK_FRAGMENT}
`

export const ComponentOwnerSidebarItem: React.FunctionComponent<{
    owner: ComponentOwnerFields['owner']
}> = ({ owner }) => (
    <>
        <ul className="list-unstyled mb-0">
            {[...(owner?.__typename === 'Group' ? owner.ancestorGroups : []), owner].filter(isDefined).map(owner => (
                <li key={owner.__typename === 'Group' ? owner.id : owner.email} className={styles.ownerPathItem}>
                    <ComponentOwnerLink owner={owner} />
                </li>
            ))}
        </ul>
        <ul className="list-unstyled d-flex flex-wrap mb-0">
            {owner?.__typename === 'Group' &&
                owner.members.map(member => (
                    <li key={member.email} className="mr-1 mb-1">
                        <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                            <UserAvatar user={member} size={19} />
                        </LinkOrSpan>
                    </li>
                ))}
        </ul>
    </>
)
