import React from 'react'

import classNames from 'classnames'

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import { getIdForLine } from './utils'

interface Props {
    selectResultFromId: (id: string) => void
    selectedResult: null | string
    result: ContentMatch
}

import styles from './FileSearchResult.module.scss'

export const FileSearchResult: React.FunctionComponent<Props> = ({
    result,
    selectedResult,
    selectResultFromId,
}: Props) => {
    const lines = result.lineMatches.map(line => {
        const key = getIdForLine(result, line)
        const onClick = (): void => selectResultFromId(key)

        return (
            // The below element's accessibility is handled via a document level event listener.
            //
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
            <div
                id={`search-result-list-item-${key}`}
                className={classNames(styles.item, {
                    [styles.itemActive]: key === selectedResult,
                })}
                onMouseDown={preventAll}
                onClick={onClick}
                key={key}
            >
                {line.line} <small>{result.path}</small>
            </div>
        )
    })

    return (
        <>
            <div className={classNames(styles.item, styles.header)}>I am a header</div>
            {lines}
        </>
    )
}

function preventAll(event: React.MouseEvent): void {
    event.stopPropagation()
    event.preventDefault()
}
