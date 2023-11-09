import React, { type FC, useCallback, useRef, useState, useEffect } from 'react'

import { mdiFileDocumentOutline, mdiFolderOutline, mdiMenuDown, mdiMenuUp } from '@mdi/js'
import classNames from 'classnames'

import { NoopEditor } from '@sourcegraph/cody-shared/dist/editor'
import { basename } from '@sourcegraph/common'
import type { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import {
    Card,
    CardHeader,
    H2,
    Icon,
    Link,
    LinkOrSpan,
    LoadingSpinner,
    ParentSize,
    StackedMeter,
    Tooltip,
    useElementObscuredArea,
} from '@sourcegraph/wildcard'

import { FileContentEditor } from '../../cody/components/FileContentEditor'
import { useCodySidebar } from '../../cody/sidebar/Provider'
import type { BlobFileFields } from '../../graphql-operations'
import { fetchBlob } from '../blob/backend'
import { RenderedFile } from '../blob/RenderedFile'

import styles from './TreePagePanels.module.scss'

interface ReadmePreviewCardProps {
    entry: TreeFields['entries'][number]
    repoName: string
    revision: string
    className?: string
}
export const ReadmePreviewCard: React.FunctionComponent<ReadmePreviewCardProps> = ({
    entry,
    repoName,
    revision,
    className,
}) => {
    const [readmeInfo, setReadmeInfo] = useState<null | BlobFileFields>(null)
    const { setEditorScope } = useCodySidebar()

    useEffect(() => {
        const subscription = fetchBlob({
            repoName,
            revision,
            filePath: entry.path,
            disableTimeout: true,
        }).subscribe(blob => {
            if (blob) {
                setReadmeInfo(blob)
            } else {
                setReadmeInfo(null)
            }
        })
        return () => subscription.unsubscribe()
    }, [repoName, revision, entry.path])

    useEffect(() => {
        if (readmeInfo) {
            setEditorScope(
                new FileContentEditor({ filePath: entry.path, repoName, revision, content: readmeInfo.content })
            )
        }

        return () => {
            if (readmeInfo) {
                setEditorScope(new NoopEditor())
            }
        }
    }, [repoName, revision, entry.path, readmeInfo, setEditorScope])

    return (
        <section className={classNames('mb-4', className)}>
            {readmeInfo ? (
                <RenderedReadmeFile blob={readmeInfo} entryUrl={entry.url} />
            ) : (
                <div className={classNames('text-muted', styles.readmeLoading)}>
                    <LoadingSpinner />
                </div>
            )}
        </section>
    )
}

interface RenderedReadmeFileProps {
    blob: BlobFileFields
    entryUrl: string
}
const RenderedReadmeFile: React.FC<RenderedReadmeFileProps> = ({ blob, entryUrl }) => {
    const renderedFileRef = useRef<HTMLDivElement>(null)
    const { bottom } = useElementObscuredArea(renderedFileRef)
    return (
        <>
            {blob.richHTML ? (
                <RenderedFile ref={renderedFileRef} dangerousInnerHTML={blob.richHTML} className={styles.readme} />
            ) : (
                <div ref={renderedFileRef} className={styles.readme}>
                    <H2 className={styles.readmePreHeader}>{basename(entryUrl)}</H2>
                    <pre className={styles.readmePre}>{blob.content}</pre>
                </div>
            )}
            {bottom > 0 && (
                <>
                    <div className={styles.readmeFader} />
                    <Link to={entryUrl} className={styles.readmeMoreLink}>
                        View full README
                    </Link>
                </>
            )}
        </>
    )
}

export interface DiffStat {
    path: string
    added: number
    deleted: number
}

export interface FilePanelProps {
    entries: Pick<TreeFields['entries'][number], 'name' | 'url' | 'isDirectory' | 'path'>[]
    diffStats?: DiffStat[]
    className?: string
    filePath: string
}

export const FilesCard: FC<FilePanelProps> = props => {
    const { entries, diffStats, className } = props

    const [sortColumn, setSortColumn] = useState<{
        column: 'Files' | 'Activity'
        direction: 'asc' | 'desc'
    }>({ column: 'Files', direction: 'asc' })

    const diffStatsByPath: { [path: string]: DiffStat } = {}
    let maxLinesChanged = 1
    if (diffStats) {
        for (const diffStat of diffStats) {
            if (diffStat.added + diffStat.deleted > maxLinesChanged) {
                maxLinesChanged = diffStat.added + diffStat.deleted
            }
            diffStatsByPath[diffStat.path] = diffStat
        }
    }

    let sortedEntries = [...entries]
    const { column, direction } = sortColumn
    switch (column) {
        case 'Files': {
            if (direction === 'desc') {
                entries.reverse()
            }
            break
        }
        case 'Activity': {
            sortedEntries = [...entries]
            if (diffStats) {
                sortedEntries.sort((entry1, entry2) => {
                    const stats1: DiffStat = diffStatsByPath[entry1.name]
                    const stats2: DiffStat = diffStatsByPath[entry2.name]
                    let difference =
                        (stats2 ? stats2.added + stats2.deleted : 0) - (stats1 ? stats1.added + stats1.deleted : 0)
                    if (direction === 'desc') {
                        difference *= -1
                    }
                    return difference
                })
            }
            break
        }
    }

    const sortCallback = useCallback(
        (column: 'Files' | 'Activity'): void => {
            if (sortColumn.column === column && sortColumn.direction === 'asc') {
                setSortColumn({ column, direction: 'desc' })
            } else {
                setSortColumn({ column, direction: 'asc' })
            }
        },
        [sortColumn]
    )
    const clickFiles = useCallback(() => sortCallback('Files'), [sortCallback])
    const keydownFiles = useCallback(
        ({ key }: React.KeyboardEvent<HTMLDivElement>) => key === 'Enter' && sortCallback('Files'),
        [sortCallback]
    )
    const clickActivity = useCallback(() => sortCallback('Activity'), [sortCallback])
    const keydownActivity = useCallback(
        ({ key }: React.KeyboardEvent<HTMLDivElement>) => key === 'Enter' && sortCallback('Activity'),
        [sortCallback]
    )

    interface Datum {
        name: 'deleted' | 'added'
        value: number
        className: string
    }

    const getDatumValue = useCallback((datum: Datum) => datum.value, [])
    const getDatumName = useCallback((datum: Datum) => datum.name, [])
    const getDatumClassName = useCallback((datum: Datum) => datum.className, [])

    return (
        <Card className={className}>
            <CardHeader className={styles.cardColHeaderWrapper}>
                <div className="container-fluid px-2">
                    <div className="row">
                        <div
                            role="button"
                            tabIndex={0}
                            onClick={clickFiles}
                            onKeyDown={keydownFiles}
                            className={classNames('d-flex flex-row align-items-start col-9 px-2', styles.cardColHeader)}
                        >
                            Files
                            <div className="flex-shrink-1 d-flex flex-column">
                                <Icon
                                    aria-label="Sort ascending"
                                    svgPath={mdiMenuUp}
                                    className={classNames(
                                        styles.sortDscIcon,
                                        sortColumn.column === 'Files' &&
                                        sortColumn.direction === 'desc' &&
                                        styles.sortSelectedIcon
                                    )}
                                />
                                <Icon
                                    aria-label="Sort descending"
                                    svgPath={mdiMenuDown}
                                    className={classNames(
                                        styles.sortAscIcon,
                                        sortColumn.column === 'Files' &&
                                        sortColumn.direction === 'asc' &&
                                        styles.sortSelectedIcon
                                    )}
                                />
                            </div>
                        </div>
                        <div
                            title="1 month activity"
                            role="button"
                            tabIndex={0}
                            onClick={clickActivity}
                            onKeyDown={keydownActivity}
                            className={classNames(
                                'd-flex flex-row-reverse align-items-start col-3 px-2 text-right',
                                styles.cardColHeader
                            )}
                        >
                            <div className="flex-shrink-1 d-flex flex-column">
                                <Icon
                                    aria-label="Sort ascending"
                                    svgPath={mdiMenuUp}
                                    className={classNames(
                                        styles.sortDscIcon,
                                        sortColumn.column === 'Activity' &&
                                        sortColumn.direction === 'desc' &&
                                        styles.sortSelectedIcon
                                    )}
                                />
                                <Icon
                                    aria-label="Sort descending"
                                    svgPath={mdiMenuDown}
                                    className={classNames(
                                        styles.sortAscIcon,
                                        sortColumn.column === 'Activity' &&
                                        sortColumn.direction === 'asc' &&
                                        styles.sortSelectedIcon
                                    )}
                                />
                            </div>
                            Recent activity
                        </div>
                    </div>
                </div>
            </CardHeader>
            <div className="container-fluid">
                {sortedEntries.map(entry => (
                    <div key={entry.name} className={classNames('row', styles.fileItem)}>
                        <div className="list-group list-group-flush px-2 py-1 col-9">
                            <LinkOrSpan
                                to={entry.url}
                                className={classNames(
                                    'test-page-file-decorable',
                                    styles.treeEntry,
                                    entry.isDirectory && 'font-weight-bold',
                                    `test-tree-entry-${entry.isDirectory ? 'directory' : 'file'}`
                                )}
                                title={entry.path}
                                data-testid="tree-entry"
                            >
                                <div
                                    className={classNames(
                                        'd-flex align-items-center justify-content-between test-file-decorable-name overflow-hidden'
                                    )}
                                >
                                    <span>
                                        <Icon
                                            className="mr-1"
                                            svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline}
                                            aria-hidden={true}
                                        />
                                        {
                                            // In case of single child expansion, we need to get the name relative to
                                            // the start of the directory (to include subdirectories)
                                        }
                                        {entry.name}
                                        {entry.isDirectory && '/'}
                                    </span>
                                </div>
                            </LinkOrSpan>
                        </div>
                        <div className="list-group list-group-flush px-2 py-1 col-3">
                            {diffStatsByPath[entry.name] && (
                                <Tooltip
                                    placement="topEnd"
                                    content={`${Intl.NumberFormat('en', {
                                        notation: 'compact',
                                    }).format(diffStatsByPath[entry.name].added)} lines added,\n${Intl.NumberFormat(
                                        'en',
                                        {
                                            notation: 'compact',
                                        }
                                    ).format(diffStatsByPath[entry.name].deleted)} lines removed in the past 30 days`}
                                >
                                    <div className={styles.meterContainer}>
                                        <ParentSize>
                                            {({ width }) => (
                                                <StackedMeter
                                                    width={width}
                                                    height={10}
                                                    viewMinMax={[0, maxLinesChanged]}
                                                    data={[
                                                        {
                                                            name: 'deleted',
                                                            value: diffStatsByPath[entry.name].deleted,
                                                            className: styles.diffStatDeleted,
                                                        },
                                                        {
                                                            name: 'added',
                                                            value: diffStatsByPath[entry.name].added,
                                                            className: styles.diffStatAdded,
                                                        },
                                                    ]}
                                                    getDatumValue={getDatumValue}
                                                    getDatumName={getDatumName}
                                                    getDatumClassName={getDatumClassName}
                                                    minBarWidth={10}
                                                    className={styles.barSvg}
                                                    rightToLeft={true}
                                                />
                                            )}
                                        </ParentSize>
                                    </div>
                                </Tooltip>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </Card>
    )
}
