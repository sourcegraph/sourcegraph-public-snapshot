import classNames from 'classnames'
import FolderIcon from 'mdi-react/FolderIcon'
import React from 'react'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ComponentSourcesFields } from '../../../../../graphql-operations'

interface Props {
    component: ComponentSourcesFields
    listGroupClassName?: string
    className?: string
}

export const ComponentSourceDefinitions: React.FunctionComponent<Props> = ({
    component: { sourceLocations },
    listGroupClassName,
    className,
}) => (
    <div className={className}>
        {sourceLocations.length > 0 ? (
            <ol className={classNames('list-group mb-0', listGroupClassName)}>
                {sourceLocations.map(sourceLocation => (
                    <li key={sourceLocation.url} className="list-group-item d-flex align-items-center py-2">
                        <FolderIcon className="icon-inline mr-1 text-muted" />
                        <RepoFileLink
                            repoName={sourceLocation.repository.name}
                            repoURL={sourceLocation.repository.url}
                            filePath={sourceLocation.path}
                            fileURL={sourceLocation.url}
                            className="d-inline"
                        />
                        {'files' in sourceLocation && sourceLocation.files && (
                            <span className="text-muted small ml-2">
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
