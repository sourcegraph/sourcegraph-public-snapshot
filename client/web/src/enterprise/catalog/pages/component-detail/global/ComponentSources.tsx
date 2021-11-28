import classNames from 'classnames'
import React from 'react'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'

import { CatalogComponentSourcesFields } from '../../../../../graphql-operations'

interface Props {
    catalogComponent: CatalogComponentSourcesFields
    className?: string
}

export const ComponentSources: React.FunctionComponent<Props> = ({
    catalogComponent: { sourceLocations },
    className,
}) =>
    sourceLocations.length > 0 ? (
        <ul className={classNames('list-unstyled', className)}>
            {sourceLocations.map(sourceLocation => (
                <li key={sourceLocation.canonicalURL}>
                    <RepoFileLink
                        repoName={sourceLocation.repository.name}
                        repoURL={sourceLocation.repository.url}
                        filePath={sourceLocation.path}
                        fileURL={sourceLocation.canonicalURL}
                    />
                </li>
            ))}
        </ul>
    ) : (
        <p className={classNames('mb-0', className)}>No source locations</p>
    )
