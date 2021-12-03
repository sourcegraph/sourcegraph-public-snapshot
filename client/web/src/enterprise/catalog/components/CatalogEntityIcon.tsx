import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import ApplicationOutlineIcon from 'mdi-react/ApplicationOutlineIcon'
import BookMultipleIcon from 'mdi-react/BookMultipleIcon'
import TextureBoxIcon from 'mdi-react/TextureBoxIcon'
import ToolsIcon from 'mdi-react/ToolsIcon'
import React from 'react'

import { CatalogComponentKind } from '../../../graphql-operations'

interface PartialEntity {
    __typename: 'CatalogComponent' | undefined
    kind: CatalogComponentKind
}

interface Props {
    entity: PartialEntity
    className?: string
}

const CATALOG_COMPONENT_ICON_BY_KIND: Record<CatalogComponentKind, React.ComponentType<{ className?: string }>> = {
    SERVICE: ApplicationCogOutlineIcon,
    WEBSITE: ApplicationOutlineIcon,
    LIBRARY: BookMultipleIcon,
    TOOL: ToolsIcon,
    OTHER: TextureBoxIcon,
}

export function catalogEntityIconComponent(entity: PartialEntity): React.ComponentType<{ className?: string }> {
    switch (entity.__typename) {
        case 'CatalogComponent':
            return CATALOG_COMPONENT_ICON_BY_KIND[entity.kind]
        default:
            return TextureBoxIcon // TODO(sqs): unexpected case
    }
}

export const CatalogEntityIcon: React.FunctionComponent<Props> = ({ entity, className }) => {
    const Icon = catalogEntityIconComponent(entity) || TextureBoxIcon
    return <Icon className={className} />
}
