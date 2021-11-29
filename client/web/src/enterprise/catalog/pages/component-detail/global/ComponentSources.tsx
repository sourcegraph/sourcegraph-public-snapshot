import path from 'path'

import classNames from 'classnames'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { getFileDecorations } from '../../../../../backend/features'
import { CatalogComponentSourcesFields } from '../../../../../graphql-operations'
import { TreeEntriesSection } from '../../../../../repo/tree/TreeEntriesSection'

import { ComponentDetailContentCardProps } from './ComponentDetailContent'

interface Props extends ComponentDetailContentCardProps, ExtensionsControllerProps, ThemeProps {
    catalogComponent: CatalogComponentSourcesFields
}

export const ComponentSources: React.FunctionComponent<Props> = ({
    catalogComponent: { sourceLocations },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
    bodyScrollableClassName,
    ...props
}) =>
    sourceLocations.length > 0 ? (
        <div className={className}>
            {false && (
                <header className={classNames('d-flex align-items-center justify-content-between', headerClassName)}>
                    <h3 className={titleClassName}>Sources</h3>
                    <Link to="TODO(sqs)" className="btn btn-link text-muted btn-sm p-0 d-flex align-items-center">
                        <SettingsIcon className="icon-inline mr-1" /> Configure
                    </Link>
                </header>
            )}
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
            <ComponentFiles
                {...props}
                sourceLocations={sourceLocations}
                className={classNames(bodyClassName, bodyScrollableClassName)}
            />
        </div>
    ) : (
        <p className={classNames('mb-0', className)}>No source locations</p>
    )

const ComponentFiles: React.FunctionComponent<
    {
        sourceLocations: CatalogComponentSourcesFields['sourceLocations']
        className?: string
    } & ExtensionsControllerProps &
        ThemeProps
> = ({ sourceLocations, className, extensionsController, isLightTheme }) => {
    const files = useMemo(
        () => sourceLocations.flatMap(sourceLocation => ('files' in sourceLocation ? sourceLocation.files : [])),
        [sourceLocations]
    )

    const fileDecorationsByPath =
        useObservable<FileDecorationsByPath>(
            useMemo(
                () =>
                    getFileDecorations({
                        files,
                        extensionsController,

                        // TODO(sqs): HACK assumes that all files are from the same repo...so hardcode it for now
                        repoName: 'github.com/sourcegraph/sourcegraph',
                        commitID: '2ada4911722e2c812cc4f1bbfb6d5d1756891392',

                        // TODO(sqs): HACK this is used for caching, this value is hacky and probably incorrect
                        parentNodeUri: sourceLocations.map(({ path }) => path).join(':'),
                    }),
                [extensionsController, files, sourceLocations]
            )
        ) ?? {}

    return (
        <ul className={classNames('list-group list-group-flush', className)}>
            {groupByParentDirectories(files).map(({ dir, files }) => (
                <li key={dir} className="list-group-item small">
                    <div className="text-muted">{dir}:</div>
                    <div className="ml-3">
                        <TreeEntriesSection
                            parentPath={dir}
                            entries={files}
                            fileDecorationsByPath={fileDecorationsByPath}
                            isLightTheme={isLightTheme}
                        />
                    </div>
                </li>
            ))}
        </ul>
    )
}

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
