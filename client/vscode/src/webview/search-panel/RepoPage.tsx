import classNames from 'classnames'
import FileDocumentOutlineIcon from 'mdi-react/FileDocumentOutlineIcon'
import FolderOutlineIcon from 'mdi-react/FolderOutlineIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { TreeEntriesVariables } from '@sourcegraph/web/src/graphql-operations'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchResults.module.scss'

import { useQueryState } from '.'

interface RepoPageProps extends WebviewPageProps {
    entries: Pick<GQL.ITreeEntry, 'name' | 'isDirectory' | 'url' | 'path'>[]
    instanceHostname: Promise<string>
    getFiles: (variables: TreeEntriesVariables) => void
    selectedRepoName: string
}

export const RepoPage: React.FunctionComponent<RepoPageProps> = ({
    entries,
    instanceHostname,
    sourcegraphVSCodeExtensionAPI,
    getFiles,
    selectedRepoName,
}) => {
    const searchActions = useQueryState(({ actions }) => actions)

    const onSelect = (isDirectory: boolean, path: string, url: string): void => {
        ;(async () => {
            const host = await instanceHostname

            switch (isDirectory) {
                case true: {
                    searchActions.setQuery({ query: `repo:^${selectedRepoName}$ file:^${path}` })
                    return getFiles({
                        repoName: selectedRepoName,
                        commitID: '',
                        revision: 'HEAD',
                        filePath: path,
                        first: 2500,
                    })
                }
                case false: {
                    return sourcegraphVSCodeExtensionAPI.openFile(`sourcegraph://${host}${url}`)
                }
            }
        })().catch(error => {
            console.log(error)
            // TODO error handling
        })
    }
    return (
        <div className={classNames('test-tree-entries mb-3', styles.section)}>
            <div
                className={classNames('pr-2', styles.treeEntriesSectionColumns, styles.treeEntriesSectionNoDecorations)}
            >
                {entries.map(entry => (
                    <Link
                        key={entry.name}
                        to={entry.url}
                        className={classNames(
                            'test-page-file-decorable',
                            styles.treeEntry,
                            entry.isDirectory && 'font-weight-bold',
                            `test-tree-entry-${entry.isDirectory ? 'directory' : 'file'}`,
                            entries.length < 7 && styles.treeEntryNoColumns
                        )}
                        title={entry.path}
                        data-testid="tree-entry"
                        onClick={() => onSelect(entry.isDirectory, entry.path, entry.url)}
                    >
                        <div
                            className={classNames(
                                'd-flex align-items-center justify-content-between test-file-decorable-name overflow-hidden'
                            )}
                        >
                            <span>
                                {entry.isDirectory && <FolderOutlineIcon className="icon-inline mr-1" />}
                                {!entry.isDirectory && <FileDocumentOutlineIcon className="icon-inline mr-1" />}
                                {entry.name}
                                {entry.isDirectory && '/'}
                            </span>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    )
}
