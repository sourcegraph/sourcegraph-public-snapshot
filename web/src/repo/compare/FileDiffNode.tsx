import { ArrowRight } from '@sourcegraph/icons/lib/ArrowRight'
import { ChevronDown } from '@sourcegraph/icons/lib/ChevronDown'
import { ChevronUp } from '@sourcegraph/icons/lib/ChevronUp'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { toBlobURL } from '../../util/url'
import { DiffStat } from './DiffStat'
import { FileDiffHunks } from './FileDiffHunks'

export interface FileDiffNodeProps {
    node: GQL.IFileDiff
    repoName: string
    base: string
    head: string
    lineNumbers: boolean
    className?: string
}

interface State {
    expanded: boolean
}

/** A file diff. */
export class FileDiffNode extends React.PureComponent<FileDiffNodeProps, State> {
    public state: State = { expanded: true }

    public render(): JSX.Element | null {
        const node = this.props.node

        const url =
            node.newPath !== null
                ? toBlobURL({ repoPath: this.props.repoName, rev: this.props.head, filePath: node.newPath })
                : toBlobURL({ repoPath: this.props.repoName, rev: this.props.base, filePath: node.oldPath! })

        return (
            <div className={`file-diff-node card ${this.props.className || ''}`}>
                <div className="card-header file-diff-node__header">
                    <div className="file-diff-node__header-path">
                        <DiffStat
                            added={node.stat.added}
                            changed={node.stat.changed}
                            deleted={node.stat.deleted}
                            className="file-diff-node__header-stat"
                        />
                        {node.newPath &&
                            (node.newPath === node.oldPath || !node.oldPath) && <code>{node.newPath}</code>}
                        {node.newPath &&
                            node.oldPath &&
                            node.newPath !== node.oldPath && (
                                <code>
                                    {node.oldPath} <ArrowRight className="icon-inline" /> {node.newPath}
                                </code>
                            )}
                        {!node.newPath && node.oldPath && <code>{node.oldPath}</code>}
                    </div>
                    <div className="file-diff-node__header-actions">
                        <Link to={url} className="btn btn-sm" data-tooltip="View file at revision">
                            View
                        </Link>
                        <button type="button" className="btn btn-sm btn-icon ml-2" onClick={this.toggleExpand}>
                            {this.state.expanded ? (
                                <ChevronDown className="icon-inline" />
                            ) : (
                                <ChevronUp className="icon-inline" />
                            )}
                        </button>
                    </div>
                </div>
                {this.state.expanded && (
                    <FileDiffHunks
                        className="file-diff-node__hunks"
                        hunks={node.hunks}
                        lineNumbers={this.props.lineNumbers}
                    />
                )}
            </div>
        )
    }

    private toggleExpand = () => this.setState(prevState => ({ expanded: !prevState.expanded }))
}
