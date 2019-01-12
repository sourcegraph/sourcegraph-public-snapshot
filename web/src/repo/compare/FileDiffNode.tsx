import { Hoverifier } from '@sourcegraph/codeintellify'
import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ActionItemProps } from '../../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'
import { DiffStat } from './DiffStat'
import { FileDiffHunks } from './FileDiffHunks'

export interface FileDiffNodeProps extends PlatformContextProps, ExtensionsControllerProps {
    node: GQL.IFileDiff

    /** The base repository and revision. */
    base: { repoName: string; repoID: GQL.ID; rev: string; commitID: string }

    /** The head repository and revision. */
    head: { repoName: string; repoID: GQL.ID; rev: string; commitID: string }

    lineNumbers: boolean
    className?: string
    location: H.Location
    history: H.History
    hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemProps>
}

interface State {
    expanded: boolean
}

/** A file diff. */
export class FileDiffNode extends React.PureComponent<FileDiffNodeProps, State> {
    public state: State = { expanded: true }

    public render(): JSX.Element | null {
        const node = this.props.node

        let path: React.ReactFragment
        if (node.newPath && (node.newPath === node.oldPath || !node.oldPath)) {
            path = <span title={node.newPath}>{node.newPath}</span>
        } else if (node.newPath && node.oldPath && node.newPath !== node.oldPath) {
            path = (
                <span title={`${node.oldPath} ⟶ ${node.newPath}`}>
                    {node.oldPath} ⟶ {node.newPath}
                </span>
            )
        } else {
            // By process of elimination (that TypeScript is unfortunately unable to infer, except
            // by reorganizing this code in a way that's much more complex to humans), node.oldPath
            // is non-null.
            path = <span title={node.oldPath!}>{node.oldPath!}</span>
        }

        const anchor = `diff-${node.internalID}`

        return (
            <>
                <a id={anchor} />
                <div className={`file-diff-node card ${this.props.className || ''}`}>
                    <div className="card-header file-diff-node__header">
                        <div className="file-diff-node__header-path-stat">
                            <DiffStat
                                added={node.stat.added}
                                changed={node.stat.changed}
                                deleted={node.stat.deleted}
                                className="file-diff-node__header-stat"
                            />
                            <Link to={{ ...this.props.location, hash: anchor }} className="file-diff-node__header-path">
                                {path}
                            </Link>
                        </div>
                        <div className="file-diff-node__header-actions">
                            <Link
                                to={node.mostRelevantFile.url}
                                className="btn btn-sm"
                                data-tooltip="View file at revision"
                            >
                                View
                            </Link>
                            <button type="button" className="btn btn-sm btn-icon ml-2" onClick={this.toggleExpand}>
                                {this.state.expanded ? (
                                    <ChevronDownIcon className="icon-inline" />
                                ) : (
                                    <ChevronUpIcon className="icon-inline" />
                                )}
                            </button>
                        </div>
                    </div>
                    {this.state.expanded && (
                        <FileDiffHunks
                            className="file-diff-node__hunks"
                            fileDiffAnchor={anchor}
                            base={{
                                ...this.props.base,
                                filePath: node.oldPath,
                            }}
                            head={{
                                ...this.props.head,
                                filePath: node.newPath,
                            }}
                            hunks={node.hunks}
                            lineNumbers={this.props.lineNumbers}
                            platformContext={this.props.platformContext}
                            history={this.props.history}
                            location={this.props.location}
                            hoverifier={this.props.hoverifier}
                        />
                    )}
                </div>
            </>
        )
    }

    private toggleExpand = () => this.setState(prevState => ({ expanded: !prevState.expanded }))
}
