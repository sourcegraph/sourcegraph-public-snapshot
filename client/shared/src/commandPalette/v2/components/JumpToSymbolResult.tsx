import React from 'react'

import { TextDocumentData } from '../../../api/viewerTypes'

interface JumpToSymbolResultProps {
    value: string
    onClick: () => void
    textDocumentData: TextDocumentData | null | undefined
}

export const JumpToSymbolResult: React.FC<JumpToSymbolResultProps> = ({ value, onClick, textDocumentData }) => {
    console.log('TODO')

    // TODO: can toggle whole-repo symbol search?
    if (!textDocumentData) {
        return (
            <div>
                <h3>Open a text document to jump to symbol</h3>
            </div>
        )
    }

    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
