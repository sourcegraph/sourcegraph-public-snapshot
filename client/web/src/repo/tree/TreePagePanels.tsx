import React, { useCallback, useRef, useState } from 'react'

import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { LinearPieChart, Link, Icon, Card, CardHeader, ParentSize } from '@sourcegraph/wildcard'

import { RenderedFile } from '../blob/RenderedFile'

import styles from './TreePage.module.scss'

interface ReadmePreviewCardProps {
    readmeHTML: string
    readmeURL: string
    location: H.Location
}

export const ReadmePreviewCard: React.FunctionComponent<ReadmePreviewCardProps> = ({
    readmeHTML,
    readmeURL,
    location,
}) => {
    const fileRef = useRef<HTMLDivElement>(null)
    const containerRef = useRef<HTMLDivElement>(null)
    const isNotCutoff =
        fileRef.current &&
        containerRef.current &&
        fileRef.current.clientHeight > 0 &&
        containerRef.current.clientHeight >= fileRef.current.clientHeight
    return (
        <>
            <div className={classNames(styles.readmeContainer)} ref={containerRef}>
                <div ref={fileRef}>
                    <RenderedFile className={styles.readme} dangerousInnerHTML={readmeHTML} location={location} />
                </div>
                <div
                    className={
                        isNotCutoff
                            ? classNames(styles.readmeFader, styles.readmeFaderInvisible)
                            : classNames(styles.readmeFader)
                    }
                />
            </div>
            <div
                className={
                    isNotCutoff
                        ? classNames(styles.readmeMore)
                        : classNames(styles.readmeMore, styles.readmeMoreVisible)
                }
            >
                <Link to={readmeURL}>More...</Link>
            </div>
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
}

export const FilesCard: React.FunctionComponent<React.PropsWithChildren<FilePanelProps>> = ({ entries, diffStats }) => {
    const [sortColumn, setSortColumn] = useState<'Files' | 'Activity'>('Files')

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

    let sortedEntries = entries
    switch (sortColumn) {
        case 'Activity':
            sortedEntries = [...entries]
            if (diffStats) {
                sortedEntries.sort((entry1, entry2) => {
                    const stats1: DiffStat = diffStatsByPath[entry1.name]
                    const stats2: DiffStat = diffStatsByPath[entry2.name]
                    return (stats2 ? stats2.added + stats2.deleted : 0) - (stats1 ? stats1.added + stats1.deleted : 0)
                })
            }
            break
    }

    const callbacks = {
        clickFiles: useCallback(() => setSortColumn('Files'), []),
        clickActivity: useCallback(() => setSortColumn('Activity'), []),
    }

    return (
        <Card className="card">
            <CardHeader>
                <div className="container-fluid px-2">
                    <div className="row">
                        <div
                            role="button"
                            tabIndex={0}
                            onClick={callbacks.clickFiles}
                            onKeyDown={callbacks.clickFiles}
                            className={classNames('col-9 px-2', styles.cardColHeader)}
                        >
                            Files
                        </div>
                        <div
                            title="1 month activity"
                            role="button"
                            tabIndex={0}
                            onClick={callbacks.clickActivity}
                            onKeyDown={callbacks.clickActivity}
                            className={classNames('col-3 px-2 text-right', styles.cardColHeader)}
                        >
                            Recent activity
                        </div>
                    </div>
                </div>
            </CardHeader>
            <div className="container-fluid">
                {sortedEntries.map(entry => (
                    <div key={entry.name} className="row">
                        <div className="list-group list-group-flush px-2 py-1 border-bottom col-9">
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
                                        {entry.name}
                                        {entry.isDirectory && '/'}
                                    </span>
                                </div>
                            </LinkOrSpan>
                        </div>
                        <div className="list-group list-group-flush px-2 py-1 border-bottom col-3">
                            {diffStatsByPath[entry.name] && (
                                <div
                                    className={styles.meterContainer}
                                    title={`+${Intl.NumberFormat('en', { notation: 'compact' }).format(
                                        diffStatsByPath[entry.name].added
                                    )}, -${Intl.NumberFormat('en', { notation: 'compact' }).format(
                                        diffStatsByPath[entry.name].deleted
                                    )} lines`}
                                >
                                    <ParentSize>
                                        {({ width }) => (
                                            <LinearPieChart
                                                width={width}
                                                height={10}
                                                viewMinMax={[0, maxLinesChanged]}
                                                data={[
                                                    {
                                                        name: 'deleted',
                                                        value: diffStatsByPath[entry.name].deleted,
                                                        color: 'red',
                                                    },
                                                    {
                                                        name: 'added',
                                                        value: diffStatsByPath[entry.name].added,
                                                        color: 'green',
                                                    },
                                                ]}
                                                getDatumValue={datum => datum.value} // TODO: factor out
                                                getDatumName={datum => datum.name}
                                                getDatumColor={datum => datum.color}
                                                minBarWidth={10}
                                                className={styles.barSvg}
                                                barRadius={5}
                                                rightToLeft={true}
                                            />
                                        )}
                                    </ParentSize>
                                </div>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </Card>
    )
}
const formatNumber = (value: number): string => Intl.NumberFormat('en', { notation: 'compact' }).format(value)

export const DiffMeter: React.FunctionComponent<{
    added: number
    deleted: number
    totalWidth: number
}> = ({ added, deleted, totalWidth }) => (
    <div title={`${formatNumber(added)} lines added, -${formatNumber(deleted)} deleted`}>
        {added > 0 && (
            <div
                // eslint-disable-next-line react/forbid-dom-props
                style={{
                    display: 'inline-block',
                    width: `${Math.max(2, (100 * added) / totalWidth)}%`,
                }}
            />
        )}
        {deleted > 0 && (
            <div
                // eslint-disable-next-line react/forbid-dom-props
                style={{
                    display: 'inline-block',
                    width: `${Math.max(2, (100 * deleted) / totalWidth)}%`,
                }}
            />
        )}
    </div>
)
