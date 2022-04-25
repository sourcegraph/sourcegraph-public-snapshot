import React, { useState } from 'react'

import classNames from 'classnames'
import AlertIcon from 'mdi-react/AlertIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined, isErrorLike } from '@sourcegraph/common'
import { Badge, Button, Link, LoadingSpinner, MenuDivider } from '@sourcegraph/wildcard'

import { RepositoryMenuContentProps } from '../../codeintel/RepositoryMenu'
import { Collapsible } from '../../components/Collapsible'
import { Timestamp } from '../../components/time/Timestamp'
import {
    LsifIndexFields,
    CodeIntelIndexerFields,
    LsifUploadFields,
    PreciseSupportLevel,
    LSIFUploadState,
    LSIFIndexState,
    SearchBasedSupportLevel,
} from '../../graphql-operations'

import { CodeIntelIndexer } from './shared/components/CodeIntelIndexer'
import { CodeIntelStateIcon } from './shared/components/CodeIntelStateIcon'
import { CodeIntelUploadOrIndexCommit } from './shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from './shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from './shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from './shared/components/CodeIntelUploadOrIndexRoot'
import {
    useCodeIntelStatus as defaultUseCodeIntelStatus,
    UseCodeIntelStatusPayload,
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from './useCodeIntelStatus'

import styles from './RepositoryMenu.module.scss'

export const RepositoryMenuContent: React.FunctionComponent<
    RepositoryMenuContentProps & {
        useCodeIntelStatus?: typeof defaultUseCodeIntelStatus
        useRequestedLanguageSupportQuery?: typeof defaultUseRequestedLanguageSupportQuery
        useRequestLanguageSupportQuery?: typeof defaultUseRequestLanguageSupportQuery
        now?: () => Date
    }
> = ({
    useCodeIntelStatus = defaultUseCodeIntelStatus,
    useRequestedLanguageSupportQuery = defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery = defaultUseRequestLanguageSupportQuery,
    now,
    ...props
}) => {
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
            <UserFacingRepositoryMenuContent
                repoName={props.repoName}
                data={data}
                useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                now={now}
            />

            {forNerds && (
                <>
                    <MenuDivider />
                    <InternalFacingRepositoryMenuContent data={data} now={now} />
                </>
            )}
        </>
    ) : null
}

//
//

const groupBy = <T, V>(values: T[], keyFn: (value: T) => V): Map<V, T[]> =>
    values.reduce(
        (map, value) => map.set(keyFn(value), (map.get(keyFn(value)) || []).concat([value])),
        new Map<V, T[]>()
    )

const getIndexerName = (uploadOrIndexer: LsifUploadFields | LsifIndexFields): string | undefined =>
    uploadOrIndexer.indexer?.name

const UserFacingRepositoryMenuContent: React.FunctionComponent<{
    repoName: string
    data: UseCodeIntelStatusPayload
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
    now?: () => Date
}> = ({ repoName, data, useRequestedLanguageSupportQuery, useRequestLanguageSupportQuery, now }) => {
    const allUploads = data.recentUploads.flatMap(uploads => uploads.uploads)
    const uploadsByIndexerName = groupBy(allUploads, getIndexerName)
    const allIndexes = data.recentIndexes.flatMap(indexes => indexes.indexes)
    const indexesByIndexerName = groupBy(allIndexes, getIndexerName)

    const nativelySupportedIndexers = (data.preciseSupport || [])
        .filter(support => support.supportLevel === PreciseSupportLevel.NATIVE)
        .map(support => support.indexers?.[0])
        .filter(isDefined)

    const allIndexers = [
        ...groupBy(
            [...allUploads, ...allIndexes]
                .map(index => index.indexer || undefined)
                .filter(isDefined)
                .concat(nativelySupportedIndexers),
            indexer => indexer.name
        ).values(),
    ].map(indexers => indexers[0])

    const languages = [
        ...new Set(
            data.searchBasedSupport
                ?.filter(support => support.supportLevel === SearchBasedSupportLevel.BASIC)
                .map(support => support.language)
        ),
    ].sort()
    const fakeIndexerNames = languages.map(name => `lsif-${name.toLowerCase()}`)
    const indexerNames = [...new Set(allIndexers.map(indexer => indexer.name).concat(fakeIndexerNames))].sort()

    // Expand badges to be as large as the maximum badge when we are displaying
    // badges of different types. This condition checks that there's at least
    // two distinct states being displayed in the following rendered component.
    const className =
        new Set(
            indexerNames.map(name =>
                (uploadsByIndexerName.get(name)?.length || 0) + (indexesByIndexerName.get(name)?.length || 0) > 0
                    ? 'enabled'
                    : allIndexers.find(candidate => candidate.name === name) !== undefined
                    ? 'configurable'
                    : 'unavailable'
            )
        ).size > 1
            ? styles.badgeMultiple
            : undefined

    return indexerNames.length === 0 ? (
        <Unsupported />
    ) : (
        <>
            {indexerNames.map((name, index) => (
                <React.Fragment key={`indexer-${name}`}>
                    {index > 0 && <MenuDivider />}
                    <IndexerSummary
                        repoName={repoName}
                        summary={{
                            name,
                            uploads: uploadsByIndexerName.get(name) || [],
                            indexes: indexesByIndexerName.get(name) || [],
                            indexer: allIndexers.find(candidate => candidate.name === name),
                        }}
                        className={className}
                        useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                        useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                        now={now}
                    />
                </React.Fragment>
            ))}
        </>
    )
}

//
//

const IndexerSummary: React.FunctionComponent<{
    repoName: string
    summary: {
        name: string
        uploads: LsifUploadFields[]
        indexes: LsifIndexFields[]
        indexer?: CodeIntelIndexerFields
    }
    className?: string
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
    now?: () => Date
}> = ({ repoName, summary, className, useRequestedLanguageSupportQuery, useRequestLanguageSupportQuery, now }) => {
    const failedUploads = summary.uploads.filter(upload => upload.state === LSIFUploadState.ERRORED)
    const failedIndexes = summary.indexes.filter(index => index.state === LSIFIndexState.ERRORED)
    const finishedAtTimes = summary.uploads.map(upload => upload.finishedAt || undefined).filter(isDefined)
    const lastUpdated = finishedAtTimes.length === 0 ? undefined : finishedAtTimes.sort().reverse()[0]

    return (
        <div className="px-2 py-1">
            <div className="d-flex align-items-center">
                <div className="px-2 py-1 text-uppercase">
                    {summary.uploads.length + summary.indexes.length > 0 ? (
                        <Badge variant="success" className={className}>
                            Enabled
                        </Badge>
                    ) : summary.indexer?.url ? (
                        <Badge variant="secondary" className={className}>
                            Configurable
                        </Badge>
                    ) : (
                        <Badge variant="outlineSecondary" className={className}>
                            Unavailable
                        </Badge>
                    )}
                </div>

                <div className="px-2 py-1">
                    <p className="mb-1">{summary.indexer?.name || summary.name} precise intelligence</p>

                    {lastUpdated && (
                        <p className="mb-1 text-muted">
                            Last updated: <Timestamp date={lastUpdated} now={now} />
                        </p>
                    )}

                    {summary.uploads.length + summary.indexes.length === 0 ? (
                        summary.indexer?.url ? (
                            <Link to={summary.indexer?.url}>Set up for this repository</Link>
                        ) : (
                            <RequestLink
                                indexerName={summary.name}
                                useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                                useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                            />
                        )
                    ) : (
                        <>
                            {failedUploads.length === 0 && failedIndexes.length === 0 && (
                                <p className="mb-1 text-muted">
                                    <CheckIcon size={16} className="text-success" /> Looks good!
                                </p>
                            )}
                            {failedUploads.length > 0 && (
                                <p className="mb-1 text-muted">
                                    <AlertIcon size={16} className="text-danger" />{' '}
                                    <Link to={`/${repoName}/-/code-intelligence/uploads?filters=errored`}>
                                        Latest upload processing
                                    </Link>{' '}
                                    failed
                                </p>
                            )}
                            {failedIndexes.length > 0 && (
                                <p className="mb-1 text-muted">
                                    <AlertIcon size={16} className="text-danger" />{' '}
                                    <Link to={`/${repoName}/-/code-intelligence/indexes?filters=errored`}>
                                        Latest indexing
                                    </Link>{' '}
                                    failed
                                </p>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}

//
//

const Unsupported: React.FunctionComponent<{}> = () => (
    <div className="px-2 py-1">
        <div className="d-flex align-items-center">
            <div className="px-2 py-1 text-uppercase">
                <Badge variant="outlineSecondary">Unsupported</Badge>
            </div>
            <div className="px-2 py-1">
                <p className="mb-0">No language detected</p>
            </div>
        </div>
    </div>
)

//
//

const InternalFacingRepositoryMenuContent: React.FunctionComponent<{
    data: UseCodeIntelStatusPayload
    now?: () => Date
}> = ({ data, now }) => {
    const preciseSupportLevels = [...new Set((data?.preciseSupport || []).map(support => support.supportLevel))].sort()
    const searchBasedSupportLevels = [
        ...new Set((data?.searchBasedSupport || []).map(support => support.supportLevel)),
    ].sort()

    return (
        <div className="px-2 py-1">
            <Collapsible titleAtStart={true} title={<h3>Activity (repo)</h3>}>
                <div>
                    <span>
                        Last auto-indexing job schedule attempt:{' '}
                        {data.lastIndexScan ? <Timestamp date={data.lastIndexScan} now={now} /> : <>never</>}
                    </span>
                </div>
                <div>
                    <span>
                        Last upload retention scan:{' '}
                        {data.lastUploadRetentionScan ? (
                            <Timestamp date={data.lastUploadRetentionScan} now={now} />
                        ) : (
                            <>never</>
                        )}
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
                                                <CodeIntelIndexer indexer={indexer} /> (
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

//
//

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

//
//

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

//
//

const RequestLink: React.FunctionComponent<{
    indexerName: string
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
}> = ({ indexerName, useRequestedLanguageSupportQuery, useRequestLanguageSupportQuery }) => {
    const language = indexerName.startsWith('lsif-') ? indexerName.slice('lsif-'.length) : indexerName

    const { data, loading: loadingSupport, error } = useRequestedLanguageSupportQuery({
        variables: {},
    })

    const [requested, setRequested] = useState(false)

    const [requestSupport, { loading: requesting, error: requestError }] = useRequestLanguageSupportQuery({
        variables: { language },
        onCompleted: () => setRequested(true),
    })

    return loadingSupport || requesting ? (
        <div className="px-2 py-1">
            <LoadingSpinner />
        </div>
    ) : error ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error loading repository summary" error={error} />
        </div>
    ) : requestError ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error requesting language support" error={requestError} />
        </div>
    ) : data ? (
        data.languages.includes(language) || requested ? (
            <span className="text-muted">
                Received your request{' '}
                <InfoCircleOutlineIcon
                    size={16}
                    data-tooltip="Requests are documented and contribute to our precise support roadmap"
                />
            </span>
        ) : (
            <Button variant="link" className={classNames('m-0 p-0', styles.languageRequest)} onClick={requestSupport}>
                I want precise support!
            </Button>
        )
    ) : null
}
