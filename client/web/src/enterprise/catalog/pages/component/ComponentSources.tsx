import path from 'path'

import classNames from 'classnames'
import React, { useMemo } from 'react'

import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { getFileDecorations } from '../../../../backend/features'
import { ComponentStateDetailFields, ComponentSourcesFields } from '../../../../graphql-operations'
import { TreeEntriesSection } from '../../../../repo/tree/TreeEntriesSection'

interface Props extends ExtensionsControllerProps, ThemeProps {
    component: ComponentSourcesFields & Pick<ComponentStateDetailFields, 'url'>
    className?: string
}

export const ComponentSources: React.FunctionComponent<Props> = ({ component, className, ...props }) => (
    <div className={className}>
        {component.sourceLocations.length > 0 && (
            <ComponentFiles {...props} sourceLocations={component.sourceLocations} />
        )}
    </div>
)

const ComponentFiles: React.FunctionComponent<
    {
        sourceLocations: ComponentSourcesFields['sourceLocations']
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
                <li key={dir} className="list-group-item small border-0">
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
