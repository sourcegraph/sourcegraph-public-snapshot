import React from 'react'

interface RecentSearchesResultProps {
    value: string
    onClick: () => void
}

export const RecentSearchesResult: React.FC<RecentSearchesResultProps> = ({ value, onClick }) => {
    console.log('TODO')
    return (
        <div>
            <h1>{value}</h1>
        </div>
    )
}
