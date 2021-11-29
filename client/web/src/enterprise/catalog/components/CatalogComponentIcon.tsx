import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import ApplicationOutlineIcon from 'mdi-react/ApplicationOutlineIcon'
import BookMultipleIcon from 'mdi-react/BookMultipleIcon'
import TextureBoxIcon from 'mdi-react/TextureBoxIcon'
import ToolsIcon from 'mdi-react/ToolsIcon'
import React from 'react'

import { CatalogComponentKind } from '../../../graphql-operations'

interface Props {
    catalogComponent: { kind: CatalogComponentKind }
    className?: string
}

export const CATALOG_COMPONENT_ICON_BY_KIND: Record<
    CatalogComponentKind,
    React.ComponentType<{ className?: string }>
> = {
    SERVICE: ApplicationCogOutlineIcon,
    WEBSITE: ApplicationOutlineIcon,
    LIBRARY: BookMultipleIcon,
    TOOL: ToolsIcon,
    OTHER: TextureBoxIcon,
}

export const CatalogComponentIcon: React.FunctionComponent<Props> = ({ catalogComponent: { kind }, className }) => {
    const Icon = CATALOG_COMPONENT_ICON_BY_KIND[kind]
    return <Icon className={className} />
}
