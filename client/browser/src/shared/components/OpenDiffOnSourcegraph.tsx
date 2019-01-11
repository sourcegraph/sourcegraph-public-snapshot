import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { IFileDiffConnection } from '../../../../../shared/src/graphql/schema'
import { queryRepositoryComparisonFileDiffs } from '../backend/diffInternalIDs'
import { OpenDiffInSourcegraphProps } from '../repo'
import { getPlatformName, repoUrlCache, sourcegraphUrl } from '../util/context'
import { Button } from './Button'

export interface Props {
    openProps: OpenDiffInSourcegraphProps
    style?: React.CSSProperties
    iconStyle?: React.CSSProperties
    className?: string
    ariaLabel?: string
    onClick?: (e: any) => void
    label: string
}

export interface State {
    fileDiff: IFileDiffConnection | undefined
}

export class OpenDiffOnSourcegraph extends React.Component<Props, State> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<Props>()

    constructor(props: Props) {
        super(props)
        this.state = { fileDiff: undefined }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    switchMap(props =>
                        queryRepositoryComparisonFileDiffs({
                            repo: this.props.openProps.repoName,
                            base: this.props.openProps.commit.baseRev,
                            head: this.props.openProps.commit.headRev,
                        })
                    ),
                    map(fileDiff => ({
                        ...fileDiff,
                        nodes: fileDiff.nodes.filter(node => node.oldPath === this.props.openProps.filePath),
                    }))
                )
                .subscribe(result => {
                    this.setState({ fileDiff: result })
                })
        )
        this.componentUpdates.next(this.props)
    }
    public render(): JSX.Element {
        const url = this.getOpenInSourcegraphUrl(this.props.openProps)
        return <Button {...this.props} className={`open-on-sourcegraph ${this.props.className}`} url={url} />
    }

    private getOpenInSourcegraphUrl(props: OpenDiffInSourcegraphProps): string {
        const baseUrl = repoUrlCache[props.repoName] || sourcegraphUrl
        const url = `${baseUrl}/${props.repoName}`
        const urlToCommit = `${url}/-/compare/${props.commit.baseRev}...${
            props.commit.headRev
        }?utm_source=${getPlatformName()}`

        if (this.state.fileDiff) {
            if (this.state.fileDiff.nodes.length > 0) {
                // Go to the specfic file in the commit diff.
                return `${urlToCommit}#diff-${this.state.fileDiff.nodes[0].internalID}`
            }
        }

        return urlToCommit
    }
}
