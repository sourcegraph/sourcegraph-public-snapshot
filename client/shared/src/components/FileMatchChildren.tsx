import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs'
import { ThemeProps } from '../theme'
import { isSettingsValid, SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { toPositionOrRangeHash, appendSubtreeQueryParameter } from '../util/url'
import { CodeExcerpt, FetchFileParameters } from './CodeExcerpt'
import { CodeExcerptUnhighlighted } from './CodeExcerptUnhighlighted'
import { FileLineMatch, MatchItem } from './FileMatch'
import { calculateMatchGroups } from './FileMatchContext'
import { Link } from './Link'
import { BadgeAttachment } from './BadgeAttachment'
import { isErrorLike } from '../util/errors'
import { ISymbol } from '../graphql/schema'
import { map } from 'rxjs/operators'

interface FileMatchProps extends SettingsCascadeProps, ThemeProps {
    location: H.Location
    items: MatchItem[]
    result: FileLineMatch
    /**
     * Whether or not to show all matches for this file, or only a subset.
     */
    allMatches: boolean
    /**
     * The number of matches to show when the results are collapsed (allMatches===false, user has not clicked "Show N more matches")
     */
    subsetMatches: number
    fetchHighlightedFileLines: (parameters: FetchFileParameters, force?: boolean) => Observable<string[]>
    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void
}

// Dev flag for disabling syntax highlighting on search results pages.
const NO_SEARCH_HIGHLIGHTING = localStorage.getItem('noSearchHighlighting') !== null

export const FileMatchChildren: React.FunctionComponent<FileMatchProps> = props => {
    const showBadges =
        props.settingsCascade.final &&
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final.experimentalFeatures &&
        // Enabled if true or null
        props.settingsCascade.final.experimentalFeatures.showBadgeAttachments !== false

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

    const maxMatches = props.allMatches ? 0 : props.subsetMatches
    const [matches, grouped] = React.useMemo(() => calculateMatchGroups(props.items, maxMatches, context), [
        props.items,
        maxMatches,
        context,
    ])

    const { result, isLightTheme, fetchHighlightedFileLines } = props
    const fetchHighlightedFileRangeLines = React.useCallback(
        (startLine, endLine) =>
            fetchHighlightedFileLines(
                {
                    repoName: result.repository.name,
                    commitID: result.file.commit.oid,
                    filePath: result.file.path,
                    disableTimeout: false,
                    isLightTheme,
                },
                false
            ).pipe(map(lines => lines.slice(startLine, endLine))),
        [result, isLightTheme, fetchHighlightedFileLines]
    )

    if (NO_SEARCH_HIGHLIGHTING) {
        return (
            <CodeExcerptUnhighlighted urlWithoutPosition={result.file.url} items={matches} onSelect={props.onSelect} />
        )
    }

    return (
        <div className="file-match-children">
            {/* Symbols */}
            {(result.symbols || []).map((symbol: ISymbol) => (
                <Link
                    to={symbol.url}
                    className="file-match-children__item test-file-match-children-item"
                    key={`symbol:${symbol.name}${String(symbol.containerName)}${symbol.url}`}
                >
                    <SymbolIcon kind={symbol.kind} className="icon-inline mr-1" />
                    <code>
                        {symbol.name}{' '}
                        {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                    </code>
                </Link>
            ))}
            {grouped.map(group => (
                <div
                    key={`linematch:${result.file.url}${group.position.line}:${group.position.character}`}
                    className="file-match-children__item-code-wrapper test-file-match-children-item-wrapper"
                >
                    <Link
                        to={appendSubtreeQueryParameter(
                            `${result.file.url}${toPositionOrRangeHash({ position: group.position })}`
                        )}
                        className="file-match-children__item file-match-children__item-clickable test-file-match-children-item"
                        onClick={props.onSelect}
                    >
                        <CodeExcerpt
                            repoName={result.repository.name}
                            commitID={result.file.commit.oid}
                            filePath={result.file.path}
                            startLine={group.startLine}
                            endLine={group.endLine}
                            highlightRanges={group.matches}
                            className="file-match-children__item-code-excerpt"
                            isLightTheme={isLightTheme}
                            fetchHighlightedFileRangeLines={fetchHighlightedFileRangeLines}
                        />
                    </Link>

                    <div className="file-match-children__item-badge-row test-badge-row">
                        {group.matches[0].badge && showBadges && (
                            // This div is necessary: it has block display, where the badge row
                            // has flex display and would cause the hover tooltip to be offset
                            // in a weird way (centered in the code context, not on the icon).
                            <div>
                                <BadgeAttachment attachment={group.matches[0].badge} isLightTheme={isLightTheme} />
                            </div>
                        )}
                    </div>
                </div>
            ))}
        </div>
    )
}
