import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { SummaryContainer } from '../../../../../../components/FilteredConnection/ui'
import { CatalogEntityForExplorerFields } from '../../../../../../graphql-operations'
import { CatalogEntityIcon } from '../../../../components/CatalogEntityIcon'
import { EntityOwner } from '../../../../components/entity-owner/EntityOwner'
import { CatalogEntityStateIndicator } from '../entity-state-indicator/EntityStateIndicator'

import styles from './CatalogExplorerList.module.scss'

interface Props {
    node: CatalogEntityForExplorerFields
    itemStartClassName?: string
    itemEndClassName?: string
    noBorder?: boolean
}

export const CatalogEntityRow: React.FunctionComponent<Props> = ({
    node,
    itemStartClassName,
    itemEndClassName,
    noBorder,
}) => (
    <>
        <h3 className={classNames('h6 font-weight-bold mb-0 d-flex align-items-center', itemStartClassName)}>
            <Link to={node.url} className={classNames('d-block text-truncate')}>
                <CatalogEntityIcon entity={node} className={classNames('icon-inline mr-1 flex-shrink-0 text-muted')} />
                {node.name}
            </Link>
            <CatalogEntityStateIndicator entity={node} className="ml-1" />
        </h3>
        <EntityOwner owner={node.owner} className="text-nowrap" blankIfNone={true} />
        <span className="text-nowrap">{node.lifecycle.toLowerCase()}</span>
        <div className={classNames('text-muted text-truncate', itemEndClassName)}>{node.description}</div>
        <div className={classNames({ 'border-top': !noBorder }, styles.separator)} />
    </>
)
