import * as React from 'react'
import { render } from 'react-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, switchMap } from 'rxjs/operators'
import { Button } from '../../shared/components/Button'
import { CodeHost, CodeHostContext } from './code_intelligence'

interface ViewOnSourcegraphButtonProps {
    context: CodeHostContext
    sourcegraphUrl: string
    ensureRepoExists: (context: CodeHostContext, sourcegraphUrl: string) => Observable<boolean>
    onConfigureSourcegraphClick?: () => void
    className?: string
}

interface ViewOnSourcegraphButtonState {
    /**
     * Whether or not the repo exists on the configured Sourcegraph instance.
     */
    repoExists?: boolean
}

class ViewOnSourcegraphButton extends React.Component<ViewOnSourcegraphButtonProps, ViewOnSourcegraphButtonState> {
    public state: ViewOnSourcegraphButtonState = {}

    private componentUpdates = new Subject<ViewOnSourcegraphButtonProps>()
    private subscriptions = new Subscription()

    constructor(props: ViewOnSourcegraphButtonProps) {
        super(props)

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(),
                    switchMap(({ context, sourcegraphUrl, ensureRepoExists }) =>
                        ensureRepoExists(context, sourcegraphUrl)
                    )
                )
                .subscribe(repoExists => {
                    this.setState({ repoExists })
                })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public render(): React.ReactNode {
        if (this.state.repoExists === undefined) {
            return null
        }

        // If repo doesn't exist and the instance is sourcegraph.com, prompt
        // user to configure Sourcegraph.
        if (
            !this.state.repoExists &&
            this.props.sourcegraphUrl === 'https://sourcegraph.com' &&
            this.props.onConfigureSourcegraphClick
        ) {
            return (
                <Button
                    label="Configure Sourcegraph"
                    onClick={this.props.onConfigureSourcegraphClick}
                    iconStyle={{ filter: 'grayscale(100%)', marginTop: '-1px', paddingRight: '4px', fontSize: '18px' }}
                    style={{ border: 'none', background: 'none' }}
                    className={`${this.props.className} btn btn-sm tooltipped tooltipped-s muted`}
                    ariaLabel="Install Sourcegraph for search and code intelligence on private repositories"
                />
            )
        }

        return (
            <Button
                url={this.getURL()}
                label="View Repository"
                ariaLabel="View repository on Sourcegraph"
                className={`open-on-sourcegraph ${this.props.className || ''}`}
            />
        )
    }

    private getURL(): string {
        const rev = this.props.context.rev ? `@${this.props.context.rev}` : ''

        return `${this.props.sourcegraphUrl}/${this.props.context.repoName}${rev}`
    }
}

export function injectViewContextOnSourcegraph(
    sourcegraphUrl: string,
    {
        getContext,
        getViewContextOnSourcegraphMount,
        contextButtonClassName,
    }: Pick<CodeHost, 'getContext' | 'getViewContextOnSourcegraphMount' | 'contextButtonClassName'>,
    ensureRepoExists: ViewOnSourcegraphButtonProps['ensureRepoExists'],
    onConfigureSourcegraphClick?: ViewOnSourcegraphButtonProps['onConfigureSourcegraphClick']
): void {
    if (!getContext || !getViewContextOnSourcegraphMount) {
        return
    }

    const mount = getViewContextOnSourcegraphMount()
    if (!mount) {
        return
    }

    render(
        <ViewOnSourcegraphButton
            context={getContext()}
            className={contextButtonClassName}
            sourcegraphUrl={sourcegraphUrl}
            ensureRepoExists={ensureRepoExists}
            onConfigureSourcegraphClick={onConfigureSourcegraphClick}
        />,
        mount
    )
}
