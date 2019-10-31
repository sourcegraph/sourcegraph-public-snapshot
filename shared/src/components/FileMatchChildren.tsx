import H from 'history'
import { flatMap } from 'lodash'
import * as React from 'react'
import { Observable } from 'rxjs'
import { ThemeProps } from '../theme'
import { isSettingsValid, SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { toPositionOrRangeHash } from '../util/url'
import { CodeExcerpt, FetchFileCtx } from './CodeExcerpt'
import { CodeExcerpt2 } from './CodeExcerpt2'
import { IFileMatch, IMatchItem } from './FileMatch'
import { mergeContext } from './FileMatchContext'
import { Link } from './Link'

interface FileMatchProps extends SettingsCascadeProps, ThemeProps {
    location: H.Location
    items: IMatchItem[]
    result: IFileMatch
    allMatches: boolean
    subsetMatches: number
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void
}

// Dev flag for disabling syntax highlighting on search results pages.
const NO_SEARCH_HIGHLIGHTING = localStorage.getItem('noSearchHighlighting') !== null

export const FileMatchChildren: React.FunctionComponent<FileMatchProps> = props => {
    const showItems = props.items
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
        .filter((item, i) => props.allMatches || i < props.subsetMatches)

    if (NO_SEARCH_HIGHLIGHTING) {
        return <CodeExcerpt2 urlWithoutPosition={props.result.file.url} items={showItems} onSelect={props.onSelect} />
    }

    // The number of lines of context to show before and after each match.
    let context = 1

    if (props.location.pathname === '/search') {
        // Check if search.contextLines is configured in settings.
        const contextLinesSetting =
            isSettingsValid(props.settingsCascade) &&
            props.settingsCascade.final &&
            props.settingsCascade.final['search.contextLines']

        if (typeof contextLinesSetting === 'number' && contextLinesSetting >= 0) {
            context = contextLinesSetting
        }
    }

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
        <div className="file-match-children">
            {/* Symbols */}
            {(props.result.symbols || []).map(symbol => (
                <Link
                    to={symbol.url}
                    className="file-match-children__item e2e-file-match-children-item"
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
                const item = items[0]
                const position = { line: item.line + 1, character: item.character + 1 }
                return (
                    <Link
                        to={`${props.result.file.url}${toPositionOrRangeHash({ position })}`}
                        key={`linematch:${props.result.file.url}${position.line}:${position.character}`}
                        className="file-match-children__item file-match-children__item-clickable e2e-file-match-children-item"
                        onClick={props.onSelect}
                    >
                        <CodeExcerpt
                            repoName={props.result.repository.name}
                            commitID={props.result.file.commit.oid}
                            filePath={props.result.file.path}
                            context={context}
                            highlightRanges={items}
                            className="file-match-children__item-code-excerpt"
                            isLightTheme={props.isLightTheme}
                            fetchHighlightedFileLines={props.fetchHighlightedFileLines}
                        />
                    </Link>
                )
            })}
        </div>
    )
}
