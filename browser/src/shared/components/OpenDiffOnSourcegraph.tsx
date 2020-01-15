import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { IFileDiffConnection } from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { queryRepositoryComparisonFileDiffs } from '../backend/diffs'
import { OpenDiffInSourcegraphProps } from '../repo'
import { getPlatformName } from '../util/context'
import { SourcegraphIconButton } from './Button'

interface Props extends PlatformContextProps<'requestGraphQL'> {
    openProps: OpenDiffInSourcegraphProps
    className?: string
    iconClassName?: string
    ariaLabel?: string
    onClick?: (e: React.MouseEvent<HTMLElement>) => void
}

interface State {
    fileDiff: IFileDiffConnection | undefined
}

export class OpenDiffOnSourcegraph extends React.Component<Props, State> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<Props>()

    constructor(props: Props) {
        super(props)
        that.state = { fileDiff: undefined }
    }

    public componentDidMount(): void {
        const { requestGraphQL } = that.props.platformContext
        that.subscriptions.add(
            // Fetch all fileDiffs in a given comparison. We rely on queryRepositoryComparisonFileDiffs
            // being memoized so that there is at most one network request when viewing
            // a commit/comparison on GitHub to get that information, despite that request occuring in
            // that component, which appears for each file in a diff.
            that.componentUpdates
                .pipe(
                    switchMap(props =>
                        queryRepositoryComparisonFileDiffs({
                            repo: that.props.openProps.repoName,
                            base: that.props.openProps.commit.baseRev,
                            head: that.props.openProps.commit.headRev,
                            requestGraphQL,
                        }).pipe(
                            map(fileDiff => ({
                                ...fileDiff,
                                // Only include the relevant file diff.
                                nodes: fileDiff.nodes.filter(node => node.oldPath === that.props.openProps.filePath),
                            })),
                            catchError(err => {
                                console.error(err)
                                return [undefined]
                            })
                        )
                    )
                )
                .subscribe(result => {
                    that.setState({ fileDiff: result })
                })
        )
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        const url = that.getOpenInSourcegraphUrl(that.props.openProps)
        return (
            <SourcegraphIconButton
                {...that.props}
                className={`open-on-sourcegraph ${that.props.className}`}
                iconClassName={that.props.iconClassName}
                url={url}
            />
        )
    }

    private getOpenInSourcegraphUrl(props: OpenDiffInSourcegraphProps): string {
        const baseUrl = props.sourcegraphURL
        const url = `${baseUrl}/${props.repoName}`
        const urlToCommit = `${url}/-/compare/${props.commit.baseRev}...${
            props.commit.headRev
        }?utm_source=${getPlatformName()}`

        if (that.state.fileDiff && that.state.fileDiff.nodes.length > 0) {
            // If the total number of files in the diff exceeds 25 (the default shown on commit pages),
            // make sure the commit page loads all files to make sure we can get to the file.
            const first =
                that.state.fileDiff.totalCount && that.state.fileDiff.totalCount > 25
                    ? `&first=${that.state.fileDiff.totalCount}`
                    : ''

            // Go to the specfic file in the commit diff using the internalID of the matched file diff.
            return `${urlToCommit}${first}#diff-${that.state.fileDiff.nodes[0].internalID}`
        }
        // If the request for fileDiffs fails, and we can't get the internal ID, just go to the comparison page.
        return urlToCommit
    }
}
