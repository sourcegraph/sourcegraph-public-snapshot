import React, { useCallback } from 'react'

import { Observable } from 'rxjs'

import { FetchBlobParameters } from '@sourcegraph/shared/src/backend/blob'
import { ResolvedRevision, ResolvedRevisionParameters } from '@sourcegraph/shared/src/backend/repo'
import { Link, LinkProps } from '@sourcegraph/wildcard'

import { BlobFileFields } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

interface FileLinkProps extends LinkProps {
    prefetch?: boolean
    resolveRevision: (parameters: ResolvedRevisionParameters) => Observable<ResolvedRevision>
    fetchBlob: (parameters: FetchBlobParameters) => Observable<BlobFileFields | null>
}

/**
 *
 */
export const FileLink: React.FunctionComponent<FileLinkProps> = ({
    prefetch,
    resolveRevision,
    fetchBlob,
    to,
    children,
    ...props
}) => {
    const prefetchFile = useCallback(async () => {
        const repo = parseBrowserRepoURL(to)
        if (repo?.filePath) {
            const revision = await resolveRevision({ repoName: repo.repoName, revision: repo.revision }).toPromise()
            await fetchBlob({
                commitID: revision.commitID,
                filePath: repo.filePath,
                repoName: repo.repoName,
                disableTimeout: false,
            }).toPromise()
        }
    }, [fetchBlob, resolveRevision, to])

    return (
        <Link to={to} onMouseEnter={prefetch ? prefetchFile : undefined} {...props}>
            {children}
        </Link>
    )
}
