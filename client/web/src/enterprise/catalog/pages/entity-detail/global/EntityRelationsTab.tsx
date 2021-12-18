import classNames from 'classnames'
import React from 'react'

import { ComponentStateDetailFields } from '../../../../../graphql-operations'

import { EntityCatalogExplorer } from './EntityCatalogExplorer'

interface Props {
    entity: Pick<ComponentStateDetailFields, 'id'>
    className?: string
}

export const EntityRelationsTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className={classNames('p-3', className)}>
        <EntityCatalogExplorer entity={entity.id} className="mb-3" />
    </div>
)
