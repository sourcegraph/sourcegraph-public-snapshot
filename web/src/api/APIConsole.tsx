import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators'

interface Props extends RouteComponentProps<any> {}

interface State {
    initialIframeSrc: string
}

export class APIConsole extends React.PureComponent<Props, State> {
    private updates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        // Precaution: Use URL fragment (not querystring) to avoid leaking sensitive querystrings in
        // HTTP referer headers.
        //
        // Also, set this initially to avoid rerendering (and destroying iframe) each time it changes.
        this.state = { initialIframeSrc: decodeURIComponent(props.location.hash) }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.updates
                .pipe(debounceTime(500))
                .subscribe(data => this.props.history.replace({ hash: encodeURIComponent(JSON.stringify(data)) }))
        )
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.updates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="api-console">
                <h2>Sourcegraph GraphQL API console</h2>
                <div className="alert alert-warning api-console__alert">
                    The API console uses your <strong>real production data</strong>.
                </div>
                <iframe
                    ref={this.setIframeRef}
                    className="api-console__iframe"
                    src={`/.api/graphql?${encodeURIComponent(this.state.initialIframeSrc)}`}
                />
            </div>
        )
    }

    private setIframeRef = (elem: HTMLIFrameElement | null) => {
        if (elem) {
            const listener = (event: MessageEvent) => {
                if (event.source === elem.contentWindow && event.origin === window.location.origin) {
                    this.updates.next(event.data)
                }
            }
            window.addEventListener('message', listener)
            this.subscriptions.add(() => window.removeEventListener('message', listener))
        }
    }
}
