import React from 'react'

interface FuzzyFinderResultProps {
    value: string
    onClick: () => void
}

export const FuzzyFinderResult: React.FC<FuzzyFinderResultProps> = ({ value, onClick }) => {
    console.log('TODO')
    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
