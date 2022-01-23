import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import {
    ComponentLabelsFields,
    ComponentDetailFields,
    ComponentTagsFields,
} from '../../../../../graphql-operations'
import { formatPersonName } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { ComponentAncestorsPath } from '../../../components/catalog-area-header/CatalogAreaHeader'
import { catalogPagePathForComponent } from '../ComponentDetailContent'
import { ComponentTag } from '../ComponentHeaderActions'

export const ComponentOwnerSidebarItem: React.FunctionComponent<{
    component: Pick<ComponentDetailFields, 'owner' | 'name' | '__typename' | 'kind' | 'url'>
    isTree?: boolean
}> = ({ component, isTree }) => (
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

export const ComponentLabelsSidebarItem: React.FunctionComponent<{
    component: ComponentLabelsFields
}> = ({ component }) =>
    component.labels.length > 0 ? (
        <dl>
            {component.labels.map(label => (
                <React.Fragment key={label.key}>
                    <dt>{label.key}</dt>
                    <dd>{label.values.join(', ')}</dd>
                </React.Fragment>
            ))}
        </dl>
    ) : null

export const ComponentTagsSidebarItem: React.FunctionComponent<{
    component: ComponentTagsFields
}> = ({ component: { tags } }) => (
    <>
        {tags.map(tag => (
            <ComponentTag
                key={tag.name}
                name={tag.name}
                components={tag.components.nodes}
                buttonClassName="p-1 border small text-muted"
            />
        ))}
    </>
)
