import React, { useEffect, useMemo } from 'react'
import { useHistory } from 'react-router'
import { Subject } from 'rxjs'
import { tap, throttleTime } from 'rxjs/operators'

import { TextDocumentData } from '../../../api/viewerTypes'
import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '../../../util/url'
import { useObservable } from '../../../util/useObservable'

import { Message } from './Message'
import { NavigableList } from './NavigableList'

interface JumpToLineResultProps {
    value: string
    onClick: () => void
    textDocumentData: TextDocumentData | null | undefined
}

// TODO: this is a web app specific implementation, abstract to platform context.
// TODO: improve performance (whole app renders?)
export const JumpToLineResult: React.FC<JumpToLineResultProps> = ({ value, onClick, textDocumentData }) => {
    const history = useHistory()

    const line = parseInt(value, 10)
    const isLineNaN = isNaN(line)

    const lines = textDocumentData?.text?.split('\n') || []
    if (lines[lines.length - 1] === '') {
        lines.splice(-1, 1)
    }
    const numberOfLines = lines.length

    const lineUpdates = useMemo(() => new Subject<{ line: number; numberOfLines: number; isLineNaN: boolean }>(), [])
    useObservable(
        useMemo(
            () =>
                lineUpdates.pipe(
                    throttleTime(150, undefined, { leading: true, trailing: true }),
                    tap(({ line, isLineNaN, numberOfLines }) => {
                        if (!isLineNaN && line <= numberOfLines) {
                            // TODO: render mode (for markdown)
                            // TODO: character
                            // TODO: abstract for bext. Disable on bext? could work by setting window.location.href

                            const searchParameters = addLineRangeQueryParameter(
                                new URLSearchParams(history.location.search),

                                toPositionOrRangeQueryParameter({
                                    position: {
                                        line,
                                    },
                                })
                            )
                            history.replace({
                                ...history.location,
                                search: searchParameters.toString(),
                            })
                        }
                    })
                ),
            []
        )
    )

    // Change line position in hash
    useEffect(() => {
        if (textDocumentData) {
            lineUpdates.next({ line, numberOfLines, isLineNaN })
        }
    }, [line, numberOfLines, isLineNaN, textDocumentData, lineUpdates])

    if (!textDocumentData) {
        return <Message type="muted">Open a text document to jump to line</Message>
    }

    if (!value || isLineNaN || line > numberOfLines) {
        return <Message type="muted">Enter a line number between 1 and {lines.length}</Message>
    }

    return (
        <NavigableList items={[null]}>
            {() => (
                <NavigableList.Item onClick={onClick} active={true}>
                    <Message>Go to line {line}</Message>
                </NavigableList.Item>
            )}
        </NavigableList>
    )
}
