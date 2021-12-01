import classNames from 'classnames'
import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogEntityOwnersFields } from '../../../../../graphql-operations'

import styles from './ComponentAuthors.module.scss'
import { EntityDetailContentCardProps } from './EntityDetailContent'

interface Props extends EntityDetailContentCardProps {
    entity: CatalogEntityOwnersFields
}

export const EntityOwners: React.FunctionComponent<Props> = ({
    entity: { owners },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
}) =>
    owners && owners.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Owners</h3>
            </header>
            <ol className={classNames('list-group list-group-horizontal border-0', bodyScrollableClassName)}>
                {owners.map(owner => (
                    <li
                        key={owner.node}
                        className={classNames(
                            'list-group-item border-top-0 border-bottom-0 text-center pt-2',
                            styles.owner
                        )}
                    >
                        {owner.node}
                        <div
                            className={classNames(styles.percent)}
                            title={`Owns ${owner.fileCount} ${pluralize('file', owner.fileCount)}`}
                        >
                            {owner.fileProportion >= 0.01 ? `${(owner.fileProportion * 100).toFixed(0)}%` : '<1%'}
                        </div>
                    </li>
                ))}
            </ol>
        </div>
    ) : (
        <div className="alert alert-info">No owners</div>
    )
