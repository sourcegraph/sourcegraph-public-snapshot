import React from 'react'

interface Props {
    className?: string
}

export const Search2Results: React.FunctionComponent<Props> = ({ className = '' }) => (
    <div className={`Search2Results ${className}`}>Results</div>
)
