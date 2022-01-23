import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { ComponentDetailFields } from '../../../../../graphql-operations'
import { formatPersonName } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { ComponentAncestorsPath } from '../../../components/catalog-area-header/CatalogAreaHeader'
import { catalogPagePathForComponent } from '../ComponentDetailContent'

export const ComponentOwnerSidebarItem: React.FunctionComponent<{
    component: Pick<ComponentDetailFields, 'owner' | 'name' | '__typename' | 'kind' | 'url'>
}> = ({ component }) => (
    <>
        <ComponentAncestorsPath
            path={catalogPagePathForComponent(component).slice(1, -1)}
            divider=">"
            className="mb-1"
            componentClassName="d-block"
            lastComponentClassName="font-weight-bold"
        />
        {/* <ComponentOwner owner={component.owner} /> */}
        <ul className="list-unstyled d-flex flex-wrap mb-0">
            {component.owner?.__typename === 'Group' &&
                component.owner.members.map(member => (
                    <li key={member.email} className="mr-1 mb-1">
                        <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                            <UserAvatar user={member} size={19} />
                        </LinkOrSpan>
                    </li>
                ))}
        </ul>
    </>
)
