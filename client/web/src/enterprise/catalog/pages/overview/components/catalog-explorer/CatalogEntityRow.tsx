import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CatalogEntityForExplorerFields, CatalogEntityRelationFields } from '../../../../../../graphql-operations'
import { CatalogEntityIcon } from '../../../../components/CatalogEntityIcon'
import { EntityOwner } from '../../../../components/entity-owner/EntityOwner'
import { catalogRelationTypeDisplayName } from '../../../../core/edges'
import { CatalogEntityStateIndicator } from '../entity-state-indicator/EntityStateIndicator'

import styles from './CatalogExplorerList.module.scss'

export interface CatalogExplorerRowStyleProps {
    itemStartClassName?: string
    itemEndClassName?: string
    noBottomBorder?: boolean
}

interface Props extends CatalogExplorerRowStyleProps {
    node: CatalogEntityForExplorerFields
    before?: string
}

export const CatalogEntityRow: React.FunctionComponent<Props> = ({
    node,
    before,
    itemStartClassName,
    itemEndClassName,
    noBottomBorder,
}) => (
    <>
        {before ? <span className={classNames('text-nowrap', itemStartClassName)}>{before}</span> : <span />}
        <h3 className={classNames('h6 font-weight-bold mb-0 d-flex align-items-center', !before && itemStartClassName)}>
            <Link to={node.url} className={classNames('d-block text-truncate')}>
                <CatalogEntityIcon entity={node} className={classNames('icon-inline mr-1 flex-shrink-0 text-muted')} />
                {node.name}
            </Link>
            <CatalogEntityStateIndicator entity={node} className="ml-1" />
        </h3>
        <EntityOwner owner={node.owner} className="text-nowrap" blankIfNone={true} />
        <span className="text-nowrap">{node.lifecycle.toLowerCase()}</span>
        <div className={classNames('text-muted text-truncate', itemEndClassName)}>{node.description}</div>
        <div className={classNames({ 'border-top': !noBottomBorder }, styles.separator)} />
    </>
)

export const CatalogEntityRowsHeader: React.FunctionComponent<
    Pick<CatalogExplorerRowStyleProps, 'itemStartClassName' | 'itemEndClassName'> & {
        before?: string
    }
> = ({ before, itemStartClassName, itemEndClassName }) => {
    const columns = [before, 'Name', 'Owner', 'Lifecycle', 'Description'].filter(isDefined)
    return (
        <>
            {!before && <span />}
            {columns.map((text, index) => (
                <div
                    key={index}
                    className={classNames(
                        'text-muted mt-2 small',
                        index === 0 && itemStartClassName,
                        index === columns.length - 1 && itemEndClassName
                    )}
                >
                    {text}
                </div>
            ))}
            <div className={classNames('border-top', styles.separator)} />
        </>
    )
}

export const CatalogEntityRelationRow: React.FunctionComponent<
    CatalogExplorerRowStyleProps & {
        edge: CatalogEntityRelationFields
    }
> = ({ edge: { node, type }, ...props }) => (
    <>
        <CatalogEntityRow {...props} node={node} before={catalogRelationTypeDisplayName(type)} />
    </>
)

export const CatalogEntityRelationRowsHeader: React.FunctionComponent<
    Omit<React.ComponentPropsWithoutRef<typeof CatalogEntityRowsHeader>, 'before'>
> = props => <CatalogEntityRowsHeader {...props} before="Relation" />
