import * as React from 'react'
import { Link } from 'react-router-dom'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { CodeExcerpt } from '../components/CodeExcerpt'
import { toPrettyBlobURL } from '../util/url'
import { ResultContainer } from './ResultContainer'

export interface IFileMatch {
    resource: string
    lineMatches: ILineMatch[]
    limitHit?: boolean
}

export interface ILineMatch {
    preview: string
    lineNumber: number
    offsetAndLengths: number[][]
    limitHit?: boolean
}

interface Props {
    /**
     * The file match search result.
     */
    result: IFileMatch

    /**
     * The icon to show left to the title.
     */
    icon: React.ComponentType<{ className: string }>

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
}

const subsetMatches = 2

export const FileMatch: React.StatelessComponent<Props> = (props: Props) => {
    const parsed = new URL(props.result.resource)
    const repoPath = parsed.hostname + parsed.pathname
    const rev = parsed.search.substr('?'.length)
    const filePath = parsed.hash.substr('#'.length)
    const items = props.result.lineMatches.map(match => ({
        range: {
            start: {
                character: match.offsetAndLengths[0][0],
                line: match.lineNumber,
            },
            end: {
                character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
                line: match.lineNumber,
            },
        },
        uri: props.result.resource,
        repoURI: repoPath,
    }))

    const title: React.ReactChild = <RepoBreadcrumb repoPath={repoPath} rev={rev} filePath={filePath} />

    const getChildren = (allMatches: boolean) => (
        <div className="file-match__list">
            {items
                .sort((a, b) => {
                    if (a.range.start.line < b.range.start.line) {
                        return -1
                    }
                    if (a.range.start.line === b.range.start.line) {
                        if (a.range.start.character < b.range.start.character) {
                            return -1
                        }
                        if (a.range.start.character === b.range.start.character) {
                            return 0
                        }
                        return 1
                    }
                    return 1
                })
                .filter((item, i) => allMatches || i < subsetMatches)
                .map((item, i) => {
                    const uri = new URL(item.uri)
                    const position = { line: item.range.start.line + 1, character: item.range.start.character + 1 }
                    return (
                        <Link
                            to={toPrettyBlobURL({
                                repoPath: uri.hostname + uri.pathname,
                                rev,
                                filePath: uri.hash.substr('#'.length),
                                position,
                            })}
                            key={i}
                            className="file-match__item"
                            onClick={props.onSelect}
                        >
                            <CodeExcerpt
                                repoPath={repoPath}
                                commitID={rev}
                                filePath={filePath}
                                position={{ line: item.range.start.line, character: item.range.start.character }}
                                highlightLength={item.range.end.character - item.range.start.character}
                                previewWindowExtraLines={1}
                                isLightTheme={props.isLightTheme}
                            />
                        </Link>
                    )
                })}
        </div>
    )

    if (props.showAllMatches) {
        return (
            <ResultContainer
                collapsible={true}
                defaultExpanded={props.expanded}
                icon={props.icon}
                title={title}
                expandedChildren={getChildren(true)}
            />
        )
    } else {
        return (
            <ResultContainer
                collapsible={items.length > subsetMatches}
                defaultExpanded={props.expanded}
                icon={props.icon}
                title={title}
                collapsedChildren={getChildren(false)}
                expandedChildren={getChildren(true)}
                collapseLabel={`Hide ${items.length - subsetMatches} matches`}
                expandLabel={`Show ${items.length - subsetMatches} matches`}
            />
        )
    }
}
