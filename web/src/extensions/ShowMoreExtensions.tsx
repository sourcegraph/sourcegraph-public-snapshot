import React from 'react'

interface Props {
    setShowMoreExtensions: React.Dispatch<React.SetStateAction<boolean>>
}

export const ShowMoreExtensions: React.FunctionComponent<Props> = ({ setShowMoreExtensions }) => (
    <button
        type="button"
        className="btn ml-2 btn-light show-more-extensions"
        onClick={() => setShowMoreExtensions(true)}
    >
        Show more extensions
    </button>
)
