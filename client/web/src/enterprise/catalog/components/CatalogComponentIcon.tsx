import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import React from 'react'

import { CatalogComponentKind } from '../../../graphql-operations'

interface Props {
    catalogComponent: { kind: CatalogComponentKind }
    className?: string
}

export const CatalogComponentIcon: React.FunctionComponent<Props> = ({ catalogComponent: { kind }, className }) => (
    <ApplicationCogOutlineIcon className={className} />
)
