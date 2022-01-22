import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import ApplicationOutlineIcon from 'mdi-react/ApplicationOutlineIcon'
import BookMultipleIcon from 'mdi-react/BookMultipleIcon'
import TextureBoxIcon from 'mdi-react/TextureBoxIcon'
import ToolsIcon from 'mdi-react/ToolsIcon'
import React from 'react'

import { ComponentKind } from '../../../graphql-operations'

interface PartialComponent {
    __typename: 'Component'
    kind: ComponentKind
}

interface Props {
    component: PartialComponent
    className?: string
}

const COMPONENT_ICON_BY_KIND: Record<ComponentKind, React.ComponentType<{ className?: string }>> = {
    SERVICE: ApplicationCogOutlineIcon,
    APPLICATION: ApplicationCogOutlineIcon,
    WEBSITE: ApplicationOutlineIcon,
    LIBRARY: BookMultipleIcon,
    TOOL: ToolsIcon,
    OTHER: TextureBoxIcon,
}

export function catalogComponentIconComponent(
    component: PartialComponent
): React.ComponentType<{ className?: string }> {
    switch (component.__typename) {
        case 'Component':
            return COMPONENT_ICON_BY_KIND[component.kind]
        default:
            return TextureBoxIcon // TODO(sqs): unexpected case
    }
}

export const CatalogComponentIcon: React.FunctionComponent<Props> = ({ component, className }) => {
    const Icon = catalogComponentIconComponent(component) || TextureBoxIcon
    return <Icon className={className} />
}
