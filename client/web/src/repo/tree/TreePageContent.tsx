import React, { useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ContributableMenu } from '@sourcegraph/client-api'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Heading, useObservable } from '@sourcegraph/wildcard'

import { getFileDecorations } from '../../backend/features'
import { TreePageRepositoryFields } from '../../graphql-operations'

import { TreeCommits } from './commits/TreeCommits'
import { TreeEntriesSection } from './TreeEntriesSection'

import styles from './TreePage.module.scss'

interface TreePageContentProps extends ExtensionsControllerProps, ThemeProps, TelemetryProps, PlatformContextProps {
    filePath: string
    tree: TreeFields
    repo: TreePageRepositoryFields
    commitID: string
    location: H.Location
    revision: string
}

export const TreePageContent: React.FunctionComponent<React.PropsWithChildren<TreePageContentProps>> = ({
    filePath,
    tree,
    repo,
    commitID,
    revision,
    ...props
}) => {
    const fileDecorationsByPath =
        useObservable<FileDecorationsByPath>(
            useMemo(
                () =>
                    getFileDecorations({
                        files: tree.entries,
                        extensionsController: props.extensionsController,
                        repoName: repo.name,
                        commitID,
                        parentNodeUri: tree.url,
                    }),
                [commitID, props.extensionsController, repo.name, tree.entries, tree.url]
            )
        ) ?? {}

    const { extensionsController } = props

    return (
        <>
            <section className={classNames('test-tree-entries mb-3', styles.section)}>
                <Heading as="h3" styleAs="h2">
                    Files and directories
                </Heading>
                <TreeEntriesSection
                    parentPath={filePath}
                    entries={tree.entries}
                    fileDecorationsByPath={fileDecorationsByPath}
                    isLightTheme={props.isLightTheme}
                />
            </section>
            {extensionsController !== null && window.context.enableLegacyExtensions ? (
                <ActionsContainer
                    {...props}
                    extensionsController={extensionsController}
                    menu={ContributableMenu.DirectoryPage}
                    empty={null}
                >
                    {items => (
                        <section className={styles.section}>
                            <Heading as="h3" styleAs="h2">
                                Actions
                            </Heading>
                            {items.map(item => (
                                <Button
                                    {...props}
                                    extensionsController={extensionsController}
                                    key={item.action.id}
                                    {...item}
                                    className="mr-1 mb-1"
                                    variant="secondary"
                                    as={ActionItem}
                                />
                            ))}
                        </section>
                    )}
                </ActionsContainer>
            ) : null}

            <TreeCommits repo={repo} commitID={commitID} filePath={filePath} className={styles.section} />
        </>
    )
}
