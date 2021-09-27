import React from 'react'

interface JumpToLineResultProps {
    value: string
    onClick: () => void
}

export const JumpToLineResult: React.FC<JumpToLineResultProps> = ({ value, onClick }) => {
    console.log('TODO')
    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
