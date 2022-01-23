import React from 'react'

import { ComponentTagsFields } from '../../../../../graphql-operations'
import { ComponentTag } from '../ComponentTag'

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
