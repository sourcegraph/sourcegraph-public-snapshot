import * as H from 'history'
import React, { useCallback, useState } from 'react'
import { VisibleChangesetApplyPreviewFields, VisibleChangesetSpecFields } from '../../../../graphql-operations'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'
import { Link } from '../../../../../../shared/src/components/Link'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { PreviewAction } from './PreviewAction'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { GitBranchChangesetDescriptionInfo } from './GitBranchChangesetDescriptionInfo'

export interface VisibleChangesetApplyPreviewNodeProps extends ThemeProps {
    node: VisibleChangesetApplyPreviewFields
    history: H.History
    location: H.Location

    /** Used for testing. **/
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. **/
    expandChangesetDescriptions?: boolean
}

export const VisibleChangesetApplyPreviewNode: React.FunctionComponent<VisibleChangesetApplyPreviewNodeProps> = ({
    node,
    isLightTheme,
    history,
    location,

    queryChangesetSpecFileDiffs,
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

    return (
        <>
            <button
                type="button"
                className="btn btn-icon test-campaigns-expand-preview d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            <PreviewAction node={node} className="visible-changeset-apply-preview-node__action" />
            <div className="visible-changeset-apply-preview-node__information">
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
                className="visible-changeset-apply-preview-node__show-details btn btn-outline-secondary d-block d-sm-none test-campaigns-expand-preview"
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
                    <div className="visible-changeset-apply-preview-node__expanded-section p-2">
                        <ExpandedSection
                            node={node}
                            history={history}
                            isLightTheme={isLightTheme}
                            location={location}
                            queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                            expandChangesetDescriptions={expandChangesetDescriptions}
                        />
                    </div>
                </>
            )}
        </>
    )
}

const ExpandedSection: React.FunctionComponent<
    {
        node: VisibleChangesetApplyPreviewFields
        history: H.History
        location: H.Location

        /** Used for testing. **/
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
        /** Expand changeset descriptions, for testing only. **/
        expandChangesetDescriptions?: boolean
    } & ThemeProps
> = ({ node, history, isLightTheme, location, queryChangesetSpecFileDiffs, expandChangesetDescriptions }) => {
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return (
            <div className="alert alert-info mb-0">
                When run, the changeset <strong>{node.targets.changeset.title}</strong> in repo{' '}
                <strong>{node.targets.changeset.repository.name}</strong> will be removed from this campaign.
            </div>
        )
    }
    if (node.targets.changesetSpec.description.__typename === 'ExistingChangesetReference') {
        return (
            <div className="alert alert-info mb-0">
                When run, the changeset with ID <strong>{node.targets.changesetSpec.description.externalID}</strong>{' '}
                will be imported from <strong>{node.targets.changesetSpec.description.baseRepository.name}</strong>.
            </div>
        )
    }
    return (
        <>
            <h4>Commits</h4>
            <GitBranchChangesetDescriptionInfo
                description={node.targets.changesetSpec.description}
                isExpandedInitially={expandChangesetDescriptions}
            />
            <h4>Diff</h4>
            <ChangesetSpecFileDiffConnection
                history={history}
                isLightTheme={isLightTheme}
                location={location}
                spec={node.targets.changesetSpec}
                queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
            />
        </>
    )
}

const ChangesetSpecFileDiffConnection: React.FunctionComponent<
    {
        spec: VisibleChangesetSpecFields
        history: H.History
        location: H.Location

        /** Used for testing. **/
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    } & ThemeProps
> = ({ spec, history, location, isLightTheme, queryChangesetSpecFileDiffs = _queryChangesetSpecFileDiffs }) => {
    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesetSpecFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                changesetSpec: spec.id,
                isLightTheme,
            }),
        [spec.id, isLightTheme, queryChangesetSpecFileDiffs]
    )
    return (
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
