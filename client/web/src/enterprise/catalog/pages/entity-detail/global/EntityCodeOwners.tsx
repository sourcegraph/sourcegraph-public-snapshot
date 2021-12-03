import classNames from 'classnames'
import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogEntityCodeOwnersFields } from '../../../../../graphql-operations'

import styles from './ComponentAuthors.module.scss'
import { EntityDetailContentCardProps } from './EntityDetailContent'

interface Props extends EntityDetailContentCardProps {
    entity: CatalogEntityCodeOwnersFields
}

export const EntityCodeOwners: React.FunctionComponent<Props> = ({
    entity: { codeOwners },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
}) =>
    codeOwners && codeOwners.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Owners</h3>
            </header>
            <ol className={classNames('list-group list-group-horizontal border-0', bodyScrollableClassName)}>
                {codeOwners.map(codeOwner => (
                    <li
                        key={codeOwner.node}
                        className={classNames(
                            'list-group-item border-top-0 border-bottom-0 text-center pt-2',
                            styles.owner
                        )}
                    >
                        {codeOwner.node}
                        <div
                            className={classNames(styles.percent)}
                            title={`Owns ${codeOwner.fileCount} ${pluralize('file', codeOwner.fileCount)}`}
                        >
                            {codeOwner.fileProportion >= 0.01
                                ? `${(codeOwner.fileProportion * 100).toFixed(0)}%`
                                : '<1%'}
                        </div>
                    </li>
                ))}
            </ol>
        </div>
    ) : (
        <div className="alert alert-info">No code owners</div>
    )
