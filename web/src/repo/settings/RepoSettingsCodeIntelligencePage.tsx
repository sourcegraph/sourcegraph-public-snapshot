import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect } from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifDumps } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Link } from '../../../../shared/src/components/Link'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'

const LsifDumpNode: FunctionComponent<{ node: GQL.ILSIFDump }> = ({ node }) => (
    <div className="w-100 lsif-dump list-group-item lsif-dump__main">
        <div className="lsif-dump__meta">
            <div className="lsif-dump__meta-root">
                <code>{node.projectRoot.commit.abbreviatedOID}</code>
                <span className="ml-2">
                    <Link to={node.projectRoot.url}>
                        <strong>{node.projectRoot.path}</strong>
                    </Link>
                </span>
            </div>
        </div>

        <small className="text-muted lsif-dump__meta-timestamp">
            <Timestamp noAbout={true} date={node.uploadedAt} />
        </small>
    </div>
)

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
}

/**
 * The repository settings code intelligence page.
 */
export const RepoSettingsCodeIntelligencePage: FunctionComponent<Props> = ({ repo, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsCodeIntelligence'), [])

    const queryDumps = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifDumps({ repository: repo.id, ...args }),
        [repo.id]
    )

    const queryLatestDumps = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifDumps({ repository: repo.id, isLatestForRepo: true, ...args }),
        [repo.id]
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence" />
            <h2>Code intelligence</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a href="https://docs.sourcegraph.com/user/code_intelligence/lsif">uploading LSIF data</a>.
            </p>

            <div className="mt-4">
                <h3>Current LSIF uploads</h3>
                <p>
                    These uploads provide code intelligence for the latest commit and are used in cross-repository{' '}
                    <em>Find References</em> requests.
                </p>

                <FilteredConnection<{}, { node: GQL.ILSIFDump }>
                    className="list-group list-group-flush mt-3"
                    noun="upload"
                    pluralNoun="uploads"
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    queryConnection={queryLatestDumps}
                    nodeComponent={LsifDumpNode}
                    history={props.history}
                    location={props.location}
                    listClassName="list-group list-group-flush"
                    cursorPaging={true}
                />
            </div>

            <div className="mt-4">
                <h3>Historic LSIF uploads</h3>
                <p>These uploads provide code intelligence for older commits.</p>

                <FilteredConnection<{}, { node: GQL.ILSIFDump }>
                    className="list-group list-group-flush mt-3"
                    noun="upload"
                    pluralNoun="uploads"
                    queryConnection={queryDumps}
                    nodeComponent={LsifDumpNode}
                    history={props.history}
                    location={props.location}
                    listClassName="list-group list-group-flush"
                    cursorPaging={true}
                />
            </div>
        </div>
    )
}
