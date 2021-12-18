import classNames from 'classnames'
import React from 'react'

import { ComponentStateDetailFields } from '../../../../../graphql-operations'

import { CatalogExplorer } from './CatalogExplorer'

interface Props {
    entity: Pick<ComponentStateDetailFields, 'id'>
    className?: string
}

export const RelationsTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className={classNames('p-3', className)}>
        <CatalogExplorer entity={entity.id} className="mb-3" />
    </div>
)
