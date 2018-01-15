import * as React from 'react'
import { Breadcrumb } from '../components/Breadcrumb'
import { toPrettyBlobURL, toTreeURL } from '../util/url'

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumb: React.SFC<{
    repoPath: string
    rev: string | undefined
    filePath: string
    isDir: boolean
}> = ({ repoPath, rev, filePath, isDir }) => {
    const parts = filePath.split('/')
    // tslint:disable-next-line:jsx-no-lambda
    return (
        <Breadcrumb
            path={filePath}
            // tslint:disable-next-line:jsx-no-lambda
            partToUrl={i => {
                const partPath = parts.slice(0, i + 1).join('/')
                if (isDir || i < parts.length - 1) {
                    return toTreeURL({ repoPath, rev, filePath: partPath })
                }
                return toPrettyBlobURL({ repoPath, rev, filePath: partPath })
            }}
            // tslint:disable-next-line:jsx-no-lambda
            partToClassName={i => (i === parts.length - 1 ? 'part-last' : 'part-directory')}
        />
    )
}
