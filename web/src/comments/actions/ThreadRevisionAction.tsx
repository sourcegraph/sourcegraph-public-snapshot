import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A repository header item that displays the branch and revision that a comment thread is attached
 * to.
 */
export const ThreadRevisionAction: React.SFC<{
    repoPath: string
    branch: string | undefined
    rev: string

    /**
     * Whether to link to the revision (should be true for repositories present on the server and
     * otherwise false).
     */
    link: boolean
}> = ({ repoPath, branch, rev, link }) => {
    const contents = `@ ${branch} (${abbreviateOID(rev)})`

    return link ? (
        <Link className="" to={`/${repoPath}@${rev}`} title={'View files at revision'}>
            {contents}
        </Link>
    ) : (
        <span>{contents}</span>
    )
}

function abbreviateOID(oid: string): string {
    if (oid.length === 40) {
        return oid.slice(0, 6)
    }
    return oid
}
