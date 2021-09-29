import React, { useEffect } from 'react'
import { useHistory } from 'react-router'

import { TextDocumentData } from '../../../api/viewerTypes'
import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '../../../util/url'

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

    // Change line position in hash
    useEffect(() => {
        if (textDocumentData) {
            if (!isLineNaN && line <= numberOfLines) {
                // TODO: render mode (for markdown)
                // TODO: character
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
        }
    }, [line, numberOfLines, isLineNaN, textDocumentData, history])

    if (!textDocumentData) {
        return (
            <div>
                <h3>Open a text document to jump to line</h3>
            </div>
        )
    }

    console.log({ lines, text: textDocumentData.text })

    // TODO: `Enter a line number between 1 and ${length}`

    // TODO: If line is not a number or it is out of range, display helpful message
    // TODO: Close on enter pressed
    return (
        <div>
            {!value && <h3>Enter a line number between 1 and {lines.length}</h3>}
            <h1>Go to line {line}</h1>
        </div>
    )
}
