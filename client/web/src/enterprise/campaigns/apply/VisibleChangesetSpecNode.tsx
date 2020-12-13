import * as H from 'history'
import React, { useCallback, useState } from 'react'
import { VisibleChangesetApplyPreviewFields } from '../../../graphql-operations'
import { ThemeProps } from '../../../../../shared/src/theme'
import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'
import { Link } from '../../../../../shared/src/components/Link'
import { DiffStat } from '../../../components/diff/DiffStat'
import { ChangesetSpecAction } from './ChangesetSpecAction'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

export interface VisibleChangesetSpecNodeProps extends ThemeProps {
    node: VisibleChangesetApplyPreviewFields
    history: H.History
    location: H.Location

    /** Used for testing. **/
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. **/
    expandChangesetDescriptions?: boolean
}

export const VisibleChangesetSpecNode: React.FunctionComponent<VisibleChangesetSpecNodeProps> = ({
    node,
    // isLightTheme,
    // history,
    // location,
    // queryChangesetSpecFileDiffs = _queryChangesetSpecFileDiffs,
    expandChangesetDescriptions = false,
}) => {
    const [isExpanded, setIsExpanded] = useState(expandChangesetDescriptions)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    /** Fetches the file diffs for the changeset */
    // const queryFileDiffs = useCallback(
    //     (args: FilteredConnectionQueryArguments) =>
    //         queryChangesetSpecFileDiffs({
    //             after: args.after ?? null,
    //             first: args.first ?? null,
    //             changesetSpec: node.id,
    //             isLightTheme,
    //         }).pipe(map(diff => diff.fileDiffs)),
    //     [node.id, isLightTheme, queryChangesetSpecFileDiffs]
    // )

    return (
        <>
            <button
                type="button"
                className="btn btn-icon test-campaigns-expand-changeset-spec d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            <ChangesetSpecAction node={node} className="visible-changeset-spec-node__action" />
            <div className="visible-changeset-spec-node__information">
                <div className="d-flex flex-column">
                    <ChangesetSpecTitle spec={node} />
                    <div className="mr-2">
                        <RepoLink spec={node} /> <References spec={node} />
                    </div>
                </div>
            </div>
            <div className="d-flex justify-content-center">
                <ApplyDiffStat spec={node} />
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <button
                type="button"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
                className="visible-changeset-spec-node__show-details btn btn-outline-secondary d-block d-sm-none test-campaigns-expand-changeset-spec"
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </button>
            {isExpanded && (
                <>
                    <div />
                    <div className="visible-changeset-spec-node__expanded-section p-2">
                        {/* {node.description.__typename === 'GitBranchChangesetDescription' && (
                            <>
                                <h4>Commits</h4>
                                <GitBranchChangesetDescriptionInfo
                                    description={node.description}
                                    isExpandedInitially={expandChangesetDescriptions}
                                />
                                <h4>Diff</h4>
                                <FileDiffConnection
                                    listClassName="list-group list-group-flush"
                                    noun="changed file"
                                    pluralNoun="changed files"
                                    queryConnection={queryFileDiffs}
                                    nodeComponent={FileDiffNode}
                                    nodeComponentProps={{
                                        history,
                                        location,
                                        isLightTheme,
                                        persistLines: true,
                                        lineNumbers: true,
                                    }}
                                    defaultFirst={15}
                                    hideSearch={true}
                                    noSummaryIfAllNodesVisible={true}
                                    history={history}
                                    location={location}
                                    useURLQuery={false}
                                    cursorPaging={true}
                                />
                            </>
                        )}
                        {node.description.__typename === 'ExistingChangesetReference' && (
                            <div className="alert alert-info mb-0">
                                When run, the changeset with ID {node.description.externalID} will be imported from{' '}
                                {node.description.baseRepository.name}.
                            </div>
                        )} */}
                    </div>
                </>
            )}
        </>
    )
}

const ChangesetSpecTitle: React.FunctionComponent<{ spec: VisibleChangesetApplyPreviewFields }> = ({ spec }) => {
    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return <h3>{spec.targets.changeset.title}</h3>
    }
    if (spec.targets.changesetSpec.description.__typename === 'ExistingChangesetReference') {
        return <h3>Import changeset #{spec.targets.changesetSpec.description.externalID}</h3>
    }
    if (
        spec.operations.length === 0 ||
        !spec.delta.titleChanged ||
        spec.targets.__typename === 'VisibleApplyPreviewTargetsAttach'
    ) {
        return <h3>{spec.targets.changesetSpec.description.title}</h3>
    }
    return (
        <h3>
            <del className="text-danger">{spec.targets.changeset.title}</del>{' '}
            <span className="text-success">{spec.targets.changesetSpec.description.title}</span>
        </h3>
    )
}

const RepoLink: React.FunctionComponent<{ spec: VisibleChangesetApplyPreviewFields }> = ({ spec }) => {
    let to: string
    let name: string
    if (
        spec.targets.__typename === 'VisibleApplyPreviewTargetsAttach' ||
        spec.targets.__typename === 'VisibleApplyPreviewTargetsUpdate'
    ) {
        to = spec.targets.changesetSpec.description.baseRepository.url
        name = spec.targets.changesetSpec.description.baseRepository.name
    } else {
        to = spec.targets.changeset.repository.url
        name = spec.targets.changeset.repository.name
    }
    return (
        <Link to={to} target="_blank" rel="noopener noreferrer" className="d-block d-sm-inline">
            {name}
        </Link>
    )
}

const References: React.FunctionComponent<{ spec: VisibleChangesetApplyPreviewFields }> = ({ spec }) => {
    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return null
    }
    if (spec.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    }
    return (
        <div className="d-block d-sm-inline-block">
            <span className="badge badge-primary">{spec.targets.changesetSpec.description.baseRef}</span> &larr;{' '}
            <span className="badge badge-primary">{spec.targets.changesetSpec.description.headRef}</span>
        </div>
    )
}

const ApplyDiffStat: React.FunctionComponent<{ spec: VisibleChangesetApplyPreviewFields }> = ({ spec }) => {
    let diffStat: { added: number; changed: number; deleted: number }
    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        if (!spec.targets.changeset.diffStat) {
            return null
        }
        diffStat = spec.targets.changeset.diffStat
    } else if (spec.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    } else {
        diffStat = spec.targets.changesetSpec.description.diffStat
    }
    return <DiffStat {...diffStat} expandedCounts={true} separateLines={true} />
}
