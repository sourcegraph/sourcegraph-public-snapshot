import React from 'react'

import { ComponentTagFields } from '../../../../../../../graphql-operations'
import { ComponentTag } from '../../../../../components/component-tag/ComponentTag'

export const ComponentTagsSidebarItem: React.FunctionComponent<{
    tags: ComponentTagFields[]
}> = ({ tags }) => (
    <>
        {tags.map(tag => (
            <ComponentTag key={tag.name} tag={tag} buttonClassName="p-1 border small text-muted" />
        ))}
    </>
)
