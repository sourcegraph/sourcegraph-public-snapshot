import classNames from 'classnames'
import React from 'react'

import { ComponentStateDetailFields } from '../../../../graphql-operations'

import { CatalogRelations } from './CatalogRelations'

interface Props {
    component: Pick<ComponentStateDetailFields, 'id'>
    className?: string
}

export const RelationsTab: React.FunctionComponent<Props> = ({ component, className }) => (
    <div className={classNames('p-3', className)}>
        <CatalogRelations component={component.id} useURLForConnectionParams={true} className="mb-3" />
    </div>
)
