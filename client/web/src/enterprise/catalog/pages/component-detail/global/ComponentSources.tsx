import path from 'path'

import classNames from 'classnames'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogComponentSourcesFields } from '../../../../../graphql-operations'
import { TreeEntriesSection } from '../../../../../repo/tree/TreeEntriesSection'

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
            <ol className={classNames('list-unstyled mb-0', bodyClassName)}>
                {sourceLocations.map(sourceLocation => (
                    <li key={sourceLocation.url} className="border p-2 m-2">
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
            <ul className={classNames('list-group list-group-flush', bodyClassName)}>
                {groupByParentDirectories(
                    sourceLocations.flatMap(sourceLocation => ('files' in sourceLocation ? sourceLocation.files : []))
                ).map(({ dir, files }) => (
                    <li key={dir} className="list-group-item small">
                        <div className="text-muted">{dir}:</div>
                        <div className="ml-3">
                            <TreeEntriesSection
                                parentPath={dir}
                                entries={files}
                                fileDecorationsByPath={{} /* TODO(sqs) */}
                                isLightTheme={false /* TODO(sqs) */}
                            />
                        </div>
                    </li>
                ))}
            </ul>
        </div>
    ) : (
        <p className={classNames('mb-0', className)}>No source locations</p>
    )

function groupByParentDirectories<F extends { path: string }>(files: F[]): { dir: string; files: F[] }[] {
    files.sort((a, b) => {
        const comp0 = path.dirname(a.path).localeCompare(path.dirname(b.path))
        return comp0 === 0 ? a.path.localeCompare(b.path) : comp0
    })

    const groups: { dir: string; files: F[] }[] = []
    for (const file of files) {
        const dirname = path.dirname(file.path)
        if (groups.length > 0 && dirname === groups[groups.length - 1].dir) {
            groups[groups.length - 1].files.push(file)
        } else {
            groups.push({ dir: dirname, files: [file] })
        }
    }

    return groups
}
