import { flatMap } from 'lodash'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { pluralize } from '../util/strings'
import { toPositionOrRangeHash } from '../util/url'
import { CodeExcerpt } from './CodeExcerpt'
import { CodeExcerpt2 } from './CodeExcerpt2'
import { mergeContext } from './FileMatchContext'
import { RepoFileLink } from './RepoFileLink'
import { Props as ResultContainerProps, ResultContainer } from './ResultContainer'

const SUBSET_COUNT_KEY = 'fileMatchSubsetCount'

export type IFileMatch = Partial<Pick<GQL.IFileMatch, 'symbols' | 'limitHit'>> & {
    file: Pick<GQL.IFile, 'path' | 'url'> & { commit: Pick<GQL.IGitCommit, 'oid'> }
    repository: Pick<GQL.IRepository, 'name' | 'url'>
    lineMatches: ILineMatch[]
}

export type ILineMatch = Pick<GQL.ILineMatch, 'preview' | 'lineNumber' | 'offsetAndLengths' | 'limitHit'>

interface IMatchItem {
    highlightRanges: {
        start: number
        highlightLength: number
    }[]
    preview: string
    line: number
}

interface Props {
    /**
     * The file match search result.
     */
    result: IFileMatch

    /**
     * The icon to show left to the title.
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this file should be rendered as expanded.
     */
    expanded: boolean

    /**
     * Whether or not to show all matches for this file, or only a subset.
     */
    showAllMatches: boolean

    isLightTheme: boolean

    allExpanded?: boolean
}

// Dev flag for disabling syntax highlighting on search results pages.
const NO_SEARCH_HIGHLIGHTING = localStorage.getItem('noSearchHighlighting') !== null

export class FileMatch extends React.PureComponent<Props> {
    private subsetMatches = 10

    constructor(props: Props) {
        super(props)

        const subsetMatches = parseInt(localStorage.getItem(SUBSET_COUNT_KEY) || '', 10)
        if (!isNaN(subsetMatches)) {
            this.subsetMatches = subsetMatches
        }
    }

    public render(): React.ReactNode {
        const result = this.props.result
        const items: IMatchItem[] = this.props.result.lineMatches.map(m => ({
            highlightRanges: m.offsetAndLengths.map(offsetAndLength => ({
                start: offsetAndLength[0],
                highlightLength: offsetAndLength[1],
            })),
            preview: m.preview,
            line: m.lineNumber,
        }))

        const title = (
            <RepoFileLink
                repoPath={result.repository.name}
                repoURL={result.repository.url}
                filePath={result.file.path}
                fileURL={result.file.url}
            />
        )

        let containerProps: ResultContainerProps

        const expandedChildren = this.getChildren(items, result, true)
        if (this.props.showAllMatches) {
            containerProps = {
                collapsible: true,
                defaultExpanded: this.props.expanded,
                icon: this.props.icon,
                title,
                expandedChildren,
                allExpanded: this.props.allExpanded,
            }
        } else {
            const len = items.length - this.subsetMatches
            containerProps = {
                collapsible: items.length > this.subsetMatches,
                defaultExpanded: this.props.expanded,
                icon: this.props.icon,
                title,
                collapsedChildren: this.getChildren(items, result, false),
                expandedChildren,
                collapseLabel: `Hide ${len} matches`,
                expandLabel: `Show ${len} more ${pluralize('match', len, 'matches')}`,
                allExpanded: this.props.allExpanded,
            }
        }

        return <ResultContainer {...containerProps} />
    }

    // If this grows any larger, it needs to be factored out into it's own component
    private getChildren = (items: IMatchItem[], result: IFileMatch, allMatches: boolean) => {
        const showItems = items
            .sort((a, b) => {
                if (a.line < b.line) {
                    return -1
                }
                if (a.line === b.line) {
                    if (a.highlightRanges[0].start < b.highlightRanges[0].start) {
                        return -1
                    }
                    if (a.highlightRanges[0].start === b.highlightRanges[0].start) {
                        return 0
                    }
                    return 1
                }
                return 1
            })
            .filter((item, i) => allMatches || i < this.subsetMatches)

        if (NO_SEARCH_HIGHLIGHTING) {
            return (
                <CodeExcerpt2 urlWithoutPosition={result.file.url} items={showItems} onSelect={this.props.onSelect} />
            )
        }

        // The number of lines of context to show before and after each match.
        const context = 1

        const groupsOfItems = mergeContext(
            context,
            flatMap(showItems, item =>
                item.highlightRanges.map(range => ({
                    line: item.line,
                    character: range.start,
                    highlightLength: range.highlightLength,
                }))
            )
        )

        return (
            <div className="file-match__list">
                {/* Symbols */}
                {(this.props.result.symbols || []).map(symbol => (
                    <Link
                        to={symbol.url}
                        className="file-match__item"
                        key={`symbol:${symbol.name}${symbol.containerName}${symbol.url}`}
                    >
                        <SymbolIcon kind={symbol.kind} className="icon-inline mr-1" />
                        <code>
                            {symbol.name}{' '}
                            {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                        </code>
                    </Link>
                ))}
                {groupsOfItems.map(items => {
                    const item = items[0]!
                    const position = { line: item.line + 1, character: item.character + 1 }
                    return (
                        <Link
                            to={`${result.file.url}${toPositionOrRangeHash({ position })}`}
                            key={`linematch:${result.file.url}${position.line}:${position.character}`}
                            className="file-match__item file-match__item-clickable"
                            onClick={this.props.onSelect}
                        >
                            <CodeExcerpt
                                repoPath={result.repository.name}
                                commitID={result.file.commit.oid}
                                filePath={result.file.path}
                                context={context}
                                highlightRanges={items}
                                className="file-match__item-code-excerpt"
                                isLightTheme={this.props.isLightTheme}
                            />
                        </Link>
                    )
                })}
            </div>
        )
    }
}
