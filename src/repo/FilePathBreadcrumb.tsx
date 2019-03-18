import * as React from 'react'
import { RepoRev } from '.'
import { LinkOrSpan } from '../components/LinkOrSpan'
import { toPrettyBlobURL, toTreeURL } from '../util/url'

interface Props {
    path: string
    partToUrl: (i: number) => string | undefined
    partToClassName?: (i: number) => string
}

/**
 * A breadcrumb where each path component is a separate link. Use this sparingly. Usually having the entire path be
 * a single link target is more usable; in that case, use RepoFileLink.
 */
class Breadcrumb extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        const parts = this.props.path.split('/')
        const spans: JSX.Element[] = []
        for (const [i, part] of parts.entries()) {
            const link = this.props.partToUrl(i)
            const className = `part ${this.props.partToClassName ? this.props.partToClassName(i) : ''} ${
                i === parts.length - 1 ? 'part-last' : ''
            }`
            spans.push(
                <LinkOrSpan key={i} className={className} to={link}>
                    {part}
                </LinkOrSpan>
            )
            if (i < parts.length - 1) {
                spans.push(
                    <span key={'sep' + i} className="breadcrumb__separator">
                        /
                    </span>
                )
            }
        }
        return (
            // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
            <span className="breadcrumb">{spans}</span>
        )
    }
}

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumb: React.SFC<
    RepoRev & {
        filePath: string
        isDir: boolean
    }
> = ({ repoPath, rev, filePath, isDir }) => {
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
