import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Link, LoadingSpinner, MenuDivider } from '@sourcegraph/wildcard'

import { RepositoryMenuContentProps } from '../../codeintel/RepositoryMenu'
import { Collapsible } from '../../components/Collapsible'
import { Timestamp } from '../../components/time/Timestamp'
import { CodeIntelIndexerFields, LsifIndexFields, LsifUploadFields } from '../../graphql-operations'

import { CodeIntelStateIcon } from './shared/components/CodeIntelStateIcon'
import { CodeIntelUploadOrIndexCommit } from './shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from './shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from './shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from './shared/components/CodeIntelUploadOrIndexRoot'
import { useCodeIntelStatus as defaultUseCodeIntelStatus, UseCodeIntelStatusPayload } from './useCodeIntelStatus'

export const RepositoryMenuContent: React.FunctionComponent<
    RepositoryMenuContentProps & {
        useCodeIntelStatus?: typeof defaultUseCodeIntelStatus
    }
> = ({ useCodeIntelStatus = defaultUseCodeIntelStatus, ...props }) => {
    const { data, loading, error } = useCodeIntelStatus({
        variables: {
            repository: props.repoName,
            commit: props.revision,
            path: props.filePath,
        },
    })

    const forNerds =
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final?.experimentalFeatures?.codeIntelRepositoryBadge?.forNerds

    return loading ? (
        <div className="px-2 py-1">
            <LoadingSpinner />
        </div>
    ) : error ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error loading repository summary" error={error} />
        </div>
    ) : data ? (
        <>
            <div className="px-2 py-1">
                <h2>Unimplemented</h2>

                <p className="text-muted">Unimplemented (enterprise version).</p>
            </div>

            {forNerds && (
                <>
                    <MenuDivider />
                    <NerdData data={data} />
                </>
            )}
        </>
    ) : null
}

const NerdData: React.FunctionComponent<{ data: UseCodeIntelStatusPayload }> = ({ data }) => {
    const preciseSupportLevels = [...new Set((data?.preciseSupport || []).map(support => support.supportLevel))].sort()
    const searchBasedSupportLevels = [
        ...new Set((data?.searchBasedSupport || []).map(support => support.supportLevel)),
    ].sort()

    return (
        <div className="px-2 py-1">
            <h2>Data for nerds</h2>

            <Collapsible titleAtStart={true} title={<h3>Activity (repo)</h3>}>
                <div>
                    <span>
                        Last auto-indexing job schedule attempt:{' '}
                        {data.lastIndexScan ? <Timestamp date={data.lastIndexScan} /> : <>never</>}
                    </span>
                </div>
                <div>
                    <span>
                        Last upload retention scan:{' '}
                        {data.lastUploadRetentionScan ? <Timestamp date={data.lastUploadRetentionScan} /> : <>never</>}
                    </span>
                </div>
            </Collapsible>

            <Collapsible titleAtStart={true} title={<h3>Support (tree)</h3>}>
                <ul>
                    {preciseSupportLevels.map(supportLevel => (
                        <li key={`precise-support-level-${supportLevel}`}>
                            <code>{supportLevel}</code>
                            <ul>
                                {data.preciseSupport
                                    ?.filter(support => support.supportLevel === supportLevel)
                                    .map(support =>
                                        support.indexers?.map(indexer => (
                                            <li key={`precise-support-level-${supportLevel}-${indexer.name}`}>
                                                <IndexerLink data={indexer} /> (
                                                {support.confidence && (
                                                    <span className="text-muted">{support.confidence}</span>
                                                )}
                                                )
                                            </li>
                                        ))
                                    )}
                            </ul>
                        </li>
                    ))}

                    {searchBasedSupportLevels.map(supportLevel => (
                        <li key={`search-support-level-${supportLevel}`}>
                            <code>{supportLevel}</code>
                            <ul>
                                {data.searchBasedSupport
                                    ?.filter(support => support.supportLevel === supportLevel)
                                    .map(support => (
                                        <li key={`search-support-level-${supportLevel}-${support.language}`}>
                                            {support.language}
                                        </li>
                                    ))}
                            </ul>
                        </li>
                    ))}
                </ul>
            </Collapsible>

            <Collapsible titleAtStart={true} title={<h3>Recent uploads (repo)</h3>}>
                <UploadOrIndexMetaTable
                    prefix="recent-uploads"
                    nodes={data.recentUploads.flatMap(namespacedUploads => namespacedUploads.uploads)}
                />
            </Collapsible>

            <Collapsible titleAtStart={true} title={<h3>Recent indexes (repo)</h3>}>
                <UploadOrIndexMetaTable
                    prefix="recent-indexes"
                    nodes={data.recentIndexes.flatMap(namespacedIndexes => namespacedIndexes.indexes)}
                />
            </Collapsible>

            <Collapsible titleAtStart={true} title={<h3>Uploads providing intel (tree)</h3>}>
                <UploadOrIndexMetaTable prefix="active-uploads" nodes={data.activeUploads} />
            </Collapsible>
        </div>
    )
}

const UploadOrIndexMetaTable: React.FunctionComponent<{
    prefix: string
    nodes: (LsifUploadFields | LsifIndexFields)[]
}> = ({ nodes, prefix }) => (
    <table className="table">
        <thead>
            <tr>
                <th>Root</th>
                <th>Commit</th>
                <th>Indexer</th>
                <th>State</th>
                <th>LastActivity</th>
            </tr>
        </thead>
        <tbody>
            {nodes.map(node => (
                <UploadOrIndexMeta key={`${prefix}-${node.id}`} data={node} />
            ))}
        </tbody>
    </table>
)

const UploadOrIndexMeta: React.FunctionComponent<{ data: LsifUploadFields | LsifIndexFields; now?: () => Date }> = ({
    data: node,
    now,
}) => (
    <tr>
        <td>
            <CodeIntelUploadOrIndexRoot node={node} />
        </td>
        <td>
            <CodeIntelUploadOrIndexCommit node={node} />
        </td>
        <td>
            <CodeIntelUploadOrIndexIndexer node={node} />
        </td>
        <td>
            <CodeIntelStateIcon state={node.state} />
        </td>
        <td>
            <CodeIntelUploadOrIndexLastActivity node={{ uploadedAt: null, queuedAt: null, ...node }} now={now} />
        </td>
    </tr>
)

const IndexerLink: React.FunctionComponent<{ data: CodeIntelIndexerFields }> = ({ data }) =>
    data.url === '' ? <>{data.name}</> : <Link to={data.url}>{data.name}</Link>
