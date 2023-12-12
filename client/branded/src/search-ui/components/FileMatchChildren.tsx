import React, { useCallback, type KeyboardEvent, type MouseEvent } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { appendLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import type { MatchGroup } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { type ContentMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { codeCopiedEvent } from '@sourcegraph/shared/src/tracking/event-log-creators'

import { CodeExcerpt } from './CodeExcerpt'
import { navigateToCodeExcerpt, navigateToFileOnMiddleMouseButtonClick } from './codeLinkNavigation'

import styles from './FileMatchChildren.module.scss'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
    result: ContentMatch
    grouped: MatchGroup[]
    /* Clicking on a match opens the link in a new tab */
    openInNewTab?: boolean
}

export const FileMatchChildren: React.FunctionComponent<React.PropsWithChildren<FileMatchProps>> = props => {
    const { result, grouped, telemetryService } = props

    const createCodeExcerptLink = (group: MatchGroup): string => {
        const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({ position: group.position })
        return appendLineRangeQueryParameter(getFileMatchUrl(result), positionOrRangeQueryParameter)
    }

    const navigate = useNavigate()
    const navigateToFile = useCallback(
        (event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>): void =>
            navigateToCodeExcerpt(event, props.openInNewTab ?? false, navigate),
        [props.openInNewTab, navigate]
    )

    const logEventOnCopy = useCallback(() => {
        telemetryService.log(...codeCopiedEvent('file-match'))
    }, [telemetryService])

    return (
        <div
            className={styles.fileMatchChildren}
            data-testid="file-match-children"
            data-selectable-search-results-group="true"
        >
            {grouped.length > 0 &&
                grouped.map(group => (
                    <div
                        key={`linematch:${getFileMatchUrl(result)}${group.position.line}:${group.position.character}`}
                        className={classNames('test-file-match-children-item-wrapper', styles.itemCodeWrapper)}
                    >
                        <div
                            data-href={createCodeExcerptLink(group)}
                            className={classNames('test-file-match-children-item', styles.item, styles.itemClickable)}
                            onClick={navigateToFile}
                            onMouseUp={navigateToFileOnMiddleMouseButtonClick}
                            onKeyDown={navigateToFile}
                            data-testid="file-match-children-item"
                            tabIndex={0}
                            role="link"
                            data-selectable-search-result="true"
                        >
                            <CodeExcerpt
                                repoName={result.repository}
                                commitID={result.commit || ''}
                                filePath={result.path}
                                startLine={group.startLine}
                                endLine={group.endLine}
                                highlightRanges={group.matches}
                                plaintextLines={group.plaintextLines}
                                highlightedLines={group.highlightedLines}
                                onCopy={logEventOnCopy}
                            />
                        </div>
                    </div>
                ))}
        </div>
    )
}
