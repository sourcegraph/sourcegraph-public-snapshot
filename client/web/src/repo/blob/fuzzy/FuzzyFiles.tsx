import { Shortcut } from '@slimsag/react-shortcuts'
import React from 'react'
import { emitKeypressEvents } from 'readline'
import { BlobInfo } from '../Blob'
import { HighlightedText, HighlightedTextProps } from './HighlightedText'
import { useLocalStorage } from './useLocalStorage'

export interface QueryProps {
    value: string
    maxResults: number
}
export interface FuzzyFilesProps {
    blobInfo: BlobInfo
    search(query: QueryProps): HighlightedTextProps[]
}

const MAX_RESULTS = 100

export const FuzzyFiles: React.FunctionComponent<FuzzyFilesProps> = props => {
    const [query, setQuery] = useLocalStorage<string>('fuzzy-files.query', '')
    const [focusIndex, setFocusIndex] = useLocalStorage<number>('fuzzy-files.focus-index', 0)
    const start = new Date()
    const candidates = props.search({
        value: query,
        maxResults: MAX_RESULTS,
    })

    // console.log(`query=${query} candidates=${candidates.map((e) => e.text)}`);
    const end = new Date()
    const elapsed = Math.max(0, end.getMilliseconds() - start.getMilliseconds())
    const visibleCandidates = candidates.slice(0, MAX_RESULTS)
    function roundFocusIndex(next: number): number {
        const modulo = next % visibleCandidates.length
        return modulo < 0 ? visibleCandidates.length + modulo : modulo
    }
    const nonVisibleCandidates = candidates.length - visibleCandidates.length
    const actualFocusIndex = roundFocusIndex(focusIndex)
    console.log('index ' + actualFocusIndex)

    return (
        <>
            <input
                id="fuzzy-files-input"
                type="text"
                onChange={e => setQuery(e.target.value)}
                value={query}
                onKeyDown={e => {
                    switch (e.key) {
                        case 'ArrowDown':
                            setFocusIndex(roundFocusIndex(focusIndex + 1))
                            break
                        case 'ArrowUp':
                            setFocusIndex(roundFocusIndex(focusIndex - 1))
                            break
                        case 'Enter':
                            const candidate = visibleCandidates[actualFocusIndex].text
                            if (candidate === props.blobInfo.filePath) break
                            const url = `/${props.blobInfo.repoName}@${props.blobInfo.commitID}/-/blob/${candidate}`
                            window.location.href = url
                            break
                        default:
                            break
                    }
                }}
            ></input>
            <p>
                {candidates.length} result{candidates.length !== 1 && 's'} in {elapsed}
                ms
                {nonVisibleCandidates > 1 && ` (only showing top ${MAX_RESULTS} results)`}
            </p>
            <ul className="fuzzy-files-results">
                {visibleCandidates.map((f, index) => {
                    const className = index === actualFocusIndex ? 'fuzzy-files-focused' : ''
                    return (
                        <li key={f.text} className={className}>
                            <HighlightedText value={f} />
                        </li>
                    )
                })}
                {nonVisibleCandidates > 0 && (
                    <li>(...{nonVisibleCandidates} hidden results, type more to narrow your filter)</li>
                )}
            </ul>
        </>
    )
}
