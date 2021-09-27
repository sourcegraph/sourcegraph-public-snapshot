import React from 'react'

interface JumpToSymbolResultProps {
    value: string
    onClick: () => void
}

export const JumpToSymbolResult: React.FC<JumpToSymbolResultProps> = ({ value, onClick }) => {
    console.log('TODO')
    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
