import React from 'react'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogComponentSourcesFields } from '../../../../../graphql-operations'

interface Props {
    catalogComponent: CatalogComponentSourcesFields
    className?: string
}

export const ComponentSourceDefinitions: React.FunctionComponent<Props> = ({
    catalogComponent: { sourceLocations },
    className,
}) => (
    <div className={className}>
        {sourceLocations.length > 0 ? (
            <ol className="list-unstyled mb-0">
                {sourceLocations.map(sourceLocation => (
                    <li key={sourceLocation.url} className="border p-2 mb-3">
                        <RepoFileLink
                            repoName={sourceLocation.repository.name}
                            repoURL={sourceLocation.repository.url}
                            filePath={sourceLocation.path}
                            fileURL={sourceLocation.url}
                            className="d-inline"
                        />{' '}
                        {'files' in sourceLocation && sourceLocation.files && (
                            <span className="text-muted small ml-1">
                                {sourceLocation.files.length} {pluralize('file', sourceLocation.files.length)}
                            </span>
                        )}
                    </li>
                ))}
            </ol>
        ) : (
            <div className="alert alert-warning">No sources defined</div>
        )}
    </div>
)
