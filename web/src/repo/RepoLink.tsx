import * as React from 'react'
import { Link } from 'react-router-dom'
import { displayRepoPath, splitPath } from '../components/RepoFileLink'

interface Props {
    repoPath: string

    /**
     * The link's destination. If null, the element is text, not a link.
     */
    to: string | null

    className?: string

    onClick?: React.MouseEventHandler<HTMLElement>
}

export const RepoLink: React.SFC<Props> = ({ repoPath, to, className, onClick }) => {
    const [repoBase, repoName] = splitPath(displayRepoPath(repoPath))
    const children = (
        <>
            {' '}
            {repoBase ? `${repoBase}/` : null}
            <strong>{repoName}</strong>
        </>
    )
    if (to === null) {
        return children
    }
    return (
        <Link className={className || ''} to={to} onClick={onClick}>
            {children}
        </Link>
    )
}
