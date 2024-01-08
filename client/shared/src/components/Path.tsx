import React from 'react'

/**
 * Displays a "/"-separated path with the last path component bolded.
 */
export const Path: React.FunctionComponent<React.PropsWithChildren<{ path: string }>> = ({ path }) => {
    if (path === '') {
        return null
    }
    const parts = path.split('/')
    return (
        <>
            {parts.length > 1 ? <span className="text-muted">{parts.slice(0, -1).join('/')}/</span> : ''}
            <strong>{parts.at(-1)}</strong>
        </>
    )
}
