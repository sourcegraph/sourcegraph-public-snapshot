import * as React from 'react'
import { Link } from 'react-router-dom'
import { SymbolHistoryEntry } from './history/utils'
import { Mode } from './RepoRevSidebarHistory'

interface HistoryEntryProps {
    symbolHistoryEntry: SymbolHistoryEntry
    index: number
    onSelect: (shiftKey: boolean, url: string, selected: boolean, index: number) => void
    prevFile: string
    mode: Mode
    selected: boolean
}

export class RepoRevSidebarHistoryEntry extends React.Component<HistoryEntryProps, {}> {
    public state = { selected: this.props.selected }

    private handleInputChange = (e: React.MouseEvent<HTMLElement>): void => {
        this.props.onSelect(e.shiftKey, this.props.symbolHistoryEntry.url, !this.props.selected, this.props.index)
    }

    public render(): JSX.Element {
        return (
            <div key={this.props.symbolHistoryEntry.url + this.props.index}>
                {this.props.symbolHistoryEntry.filePath !== this.props.prevFile && (
                    <div className="repo-rev-sidebar-history__header">
                        <div className="repo-rev-sidebar-history__text text-muted small">
                            {this.props.symbolHistoryEntry.repoPath}
                        </div>
                        <div className="repo-rev-sidebar-history__text"> {this.props.symbolHistoryEntry.filePath}</div>
                    </div>
                )}
                <div
                    key={this.props.symbolHistoryEntry.url + this.props.index + ' separator'}
                    className=" list-group-item"
                >
                    <Link
                        to={this.props.symbolHistoryEntry.url}
                        key={this.props.symbolHistoryEntry.url + this.props.index}
                        className="repo-rev-sidebar-history__link"
                    >
                        <li className="repo-rev-sidebar-history__list-item-content">
                            {this.props.mode === 'DOC' ? (
                                <>
                                    <div
                                        className="repo-rev-sidebar-history__symbol-title repo-rev-sidebar-history__text"
                                        dangerouslySetInnerHTML={{ __html: this.props.symbolHistoryEntry.name }}
                                    />
                                    <div className="repo-rev-sidebar-history__text text-muted small">{`
                                            ${this.props.symbolHistoryEntry.repoPath} - ${
                                        this.props.symbolHistoryEntry.filePath
                                    } ${this.props.symbolHistoryEntry.lineNumber &&
                                        `- L${this.props.symbolHistoryEntry.lineNumber}`}`}</div>
                                    {this.props.symbolHistoryEntry.hoverContents[0] && (
                                        <div
                                            className="repo-rev-sidebar-history__contents"
                                            dangerouslySetInnerHTML={{
                                                __html: this.props.symbolHistoryEntry.hoverContents[0],
                                            }}
                                        />
                                    )}
                                </>
                            ) : (
                                <div className="repo-rev-sidebar-history__code">
                                    {this.props.symbolHistoryEntry.linesOfCode &&
                                        this.props.symbolHistoryEntry.linesOfCode.map((line, i) => (
                                            <div
                                                className={
                                                    this.props.symbolHistoryEntry.lineNumber === 1 ||
                                                    this.props.symbolHistoryEntry.lineNumber === 2
                                                        ? i === this.props.symbolHistoryEntry.lineNumber - 1
                                                            ? 'selection-highlight'
                                                            : ''
                                                        : i === 2
                                                            ? 'selection-highlight'
                                                            : ''
                                                }
                                                dangerouslySetInnerHTML={{ __html: line }}
                                                key={line + i}
                                            />
                                        ))}
                                </div>
                            )}
                        </li>
                    </Link>
                    <div className="repo-rev-sidebar-history__checkbox">
                        <input type="checkbox" onClick={this.handleInputChange} checked={this.props.selected} />
                    </div>
                </div>
            </div>
        )
    }
}
