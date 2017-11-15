import * as React from 'react'
import { Link } from 'react-router-dom'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { CodeExcerpt } from '../components/CodeExcerpt'
import { toPrettyBlobURL } from '../util/url'
import { ResultContainer } from './ResultContainer'

interface Props {
    /**
     * The file match search result.
     */
    result: GQL.IFileMatch

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
}

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

    return (
        <ResultContainer collapsible={true} defaultExpanded={props.expanded} icon={props.icon} title={title}>
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
                                />
                            </Link>
                        )
                    })}
            </div>
        </ResultContainer>
    )
}
