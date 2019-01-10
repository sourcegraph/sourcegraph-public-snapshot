import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { IFileDiffConnection } from '../../../../../shared/src/graphql/schema'
import { queryRepositoryComparisonFileDiffs } from '../backend/diffInternalIDs'
import { OpenInSourcegraphProps } from '../repo'
import { getPlatformName, repoUrlCache, sourcegraphUrl } from '../util/context'
import { Button } from './Button'

export interface Props {
    openProps: OpenInSourcegraphProps
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

export class OpenDiffOnSourcegraph extends React.Component<Props, {}> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<Props>()

    constructor(props: Props) {
        super(props)
        this.state = { fileDiff: undefined }
    }

    public componentDidMount(): void {
        console.log('PROPS', this.props)
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    switchMap(props =>
                        queryRepositoryComparisonFileDiffs({
                            repo: this.props.openProps.repoName,
                            base: this.props.openProps.commit!.baseRev,
                            head: this.props.openProps.commit!.headRev,
                        })
                    )
                )
                .subscribe(result => {
                    console.log('setting state!')
                    this.setState({ fileDiff: result })
                })
        )
        this.componentUpdates.next(this.props)
    }
    public render(): JSX.Element {
        console.log('STATE', this.state)
        const url = this.getOpenInSourcegraphUrl(this.props.openProps)
        return <Button {...this.props} className={`open-on-sourcegraph ${this.props.className}`} url={url} />
    }

    private getOpenInSourcegraphUrl(props: OpenInSourcegraphProps): string {
        const baseUrl = repoUrlCache[props.repoName] || sourcegraphUrl
        // Build URL for Web
        let url = `${baseUrl}/${props.repoName}`
        if (props.commit) {
            return `${url}/-/compare/${props.commit.baseRev}...${props.commit.headRev}?utm_source=${getPlatformName()}`
        }
        if (props.rev) {
            url = `${url}@${props.rev}`
        }
        if (props.filePath) {
            url = `${url}/-/blob/${props.filePath}`
        }
        if (props.query) {
            if (props.query.diff) {
                url = `${url}?diff=${props.query.diff.rev}&utm_source=${getPlatformName()}`
            } else if (props.query.search) {
                url = `${url}?q=${props.query.search}&utm_source=${getPlatformName()}`
            }
        }
        if (props.coords) {
            url = `${url}#L${props.coords.line}:${props.coords.char}`
        }
        if (props.fragment) {
            url = `${url}$${props.fragment}`
        }
        return url
    }
}
