import classNames from 'classnames'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'

import { CatalogComponentSourcesFields } from '../../../../../graphql-operations'

interface Props {
    catalogComponent: CatalogComponentSourcesFields
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
}

export const ComponentSources: React.FunctionComponent<Props> = ({
    catalogComponent: { sourceLocations },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
}) =>
    sourceLocations.length > 0 ? (
        <div className={className}>
            <header className={classNames('d-flex align-items-center justify-content-between', headerClassName)}>
                <h3 className={titleClassName}>Sources</h3>
                <Link to="TODO(sqs)" className="btn btn-link text-muted btn-sm p-0 d-flex align-items-center">
                    <SettingsIcon className="icon-inline mr-1" /> Configure
                </Link>
            </header>
            <ol className={classNames('list-group list-group-flush', bodyClassName)}>
                {sourceLocations.map(sourceLocation => (
                    <li key={sourceLocation.canonicalURL} className="list-group-item">
                        <RepoFileLink
                            repoName={sourceLocation.repository.name}
                            repoURL={sourceLocation.repository.url}
                            filePath={sourceLocation.path}
                            fileURL={sourceLocation.canonicalURL}
                        />
                    </li>
                ))}
            </ol>
            <p className="card-body">All files</p>
        </div>
    ) : (
        <p className={classNames('mb-0', className)}>No source locations</p>
    )
