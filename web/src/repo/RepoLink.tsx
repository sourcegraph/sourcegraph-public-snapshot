import * as React from 'react'
import { Link } from 'react-router-dom'
import { displayRepoPath, splitPath } from '../components/RepoFileLink'

interface Props {
    repoPath: string
    rev?: string

    /**
     * The link's destination, if the caller wants to override the default destination of the repository's root at
     * the given revision. If null, the element is text, not a link.
     */
    to?: string | null

    className?: string

    onClick?: React.MouseEventHandler<HTMLElement>
}

const NotALink: React.SFC<{ to: any; children: React.ReactElement<any> }> = ({ children }) => children || null

export const RepoLink: React.SFC<Props> = ({ repoPath, rev, to, className, onClick }) => {
    const L = to === null ? NotALink : Link

    const [repoBase, repoName] = splitPath(displayRepoPath(repoPath))
    return (
        <L
            className={className || ''}
            to={typeof to === 'string' ? to : `/${repoPath}${rev ? `@${rev}` : ''}`}
            onClick={onClick}
        >
            {repoBase ? `${repoBase}/` : null}
            <strong>{repoName}</strong>
        </L>
    )
}
