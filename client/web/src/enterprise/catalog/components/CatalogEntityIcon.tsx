import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import ApplicationOutlineIcon from 'mdi-react/ApplicationOutlineIcon'
import BookMultipleIcon from 'mdi-react/BookMultipleIcon'
import PackageIcon from 'mdi-react/PackageIcon'
import TextureBoxIcon from 'mdi-react/TextureBoxIcon'
import ToolsIcon from 'mdi-react/ToolsIcon'
import React from 'react'

import { ComponentKind } from '../../../graphql-operations'

type PartialEntity =
    | {
          __typename: 'Component'
          kind: ComponentKind
      }
    | { __typename: 'Package' }

interface Props {
    entity: PartialEntity
    className?: string
}

const COMPONENT_ICON_BY_KIND: Record<ComponentKind, React.ComponentType<{ className?: string }>> = {
    SERVICE: ApplicationCogOutlineIcon,
    WEBSITE: ApplicationOutlineIcon,
    LIBRARY: BookMultipleIcon,
    TOOL: ToolsIcon,
    OTHER: TextureBoxIcon,
}

export function componentIconComponent(entity: PartialEntity): React.ComponentType<{ className?: string }> {
    switch (entity.__typename) {
        case 'Component':
            return COMPONENT_ICON_BY_KIND[entity.kind]
        case 'Package':
            return PackageIcon
        default:
            return TextureBoxIcon // TODO(sqs): unexpected case
    }
}

export const ComponentIcon: React.FunctionComponent<Props> = ({ entity, className }) => {
    const Icon = componentIconComponent(entity) || TextureBoxIcon
    return <Icon className={className} />
}
