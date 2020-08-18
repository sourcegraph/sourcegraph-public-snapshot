import React from 'react'

interface Props {
    setShowMoreExtensions: React.Dispatch<React.SetStateAction<boolean>>
}

export const ShowMoreExtensions: React.FunctionComponent<Props> = ({ setShowMoreExtensions }) => (
    <div className="show-more-extensions__container">
    <button type="button" className="btn show-more-extensions" onClick={() => setShowMoreExtensions(true)}>
        Show more extensions
    </button>
    </div>
)
