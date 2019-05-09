import classNames from 'classnames'
import { isEqual } from 'lodash'
import * as React from 'react'
import { render } from 'react-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { SourcegraphIconButton } from '../../shared/components/Button'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { CodeHost, CodeHostContext } from './code_intelligence'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps extends ViewOnSourcegraphButtonClassProps {
    context: CodeHostContext
    sourcegraphURL: string
    ensureRepoExists: (context: CodeHostContext, sourcegraphUrl: string) => Observable<boolean>
    onConfigureSourcegraphClick?: () => void
}

interface ViewOnSourcegraphButtonState {
    /**
     * Whether or not the repo exists on the configured Sourcegraph instance.
     */
    repoExists?: boolean
}

class ViewOnSourcegraphButton extends React.Component<ViewOnSourcegraphButtonProps, ViewOnSourcegraphButtonState> {
    private componentUpdates = new Subject<ViewOnSourcegraphButtonProps>()
    private subscriptions = new Subscription()

    constructor(props: ViewOnSourcegraphButtonProps) {
        super(props)
        this.state = {}
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ context, sourcegraphURL, ensureRepoExists }) => ({
                        context,
                        sourcegraphURL,
                        ensureRepoExists,
                    })),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(({ context, sourcegraphURL, ensureRepoExists }) =>
                        ensureRepoExists(context, sourcegraphURL)
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

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        if (this.state.repoExists === undefined) {
            return null
        }

        // If repo doesn't exist and the instance is sourcegraph.com, prompt
        // user to configure Sourcegraph.
        if (
            !this.state.repoExists &&
            this.props.sourcegraphURL === DEFAULT_SOURCEGRAPH_URL &&
            this.props.onConfigureSourcegraphClick
        ) {
            return (
                <SourcegraphIconButton
                    label="Configure Sourcegraph"
                    ariaLabel="Install Sourcegraph for search and code intelligence on private instance"
                    className={classNames('open-on-sourcegraph', this.props.className)}
                    iconClassName={classNames('open-on-sourcegraph__icon--muted', this.props.iconClassName)}
                    onClick={this.props.onConfigureSourcegraphClick}
                />
            )
        }

        return (
            <SourcegraphIconButton
                url={this.getURL()}
                ariaLabel="View repository on Sourcegraph"
                className={classNames('open-on-sourcegraph', this.props.className)}
                iconClassName={this.props.iconClassName}
            />
        )
    }

    private getURL(): string {
        const rev = this.props.context.rev ? `@${this.props.context.rev}` : ''

        return `${this.props.sourcegraphURL}/${this.props.context.repoName}${rev}`
    }
}

export const renderViewContextOnSourcegraph = ({
    sourcegraphURL,
    getContext,
    ensureRepoExists,
    viewOnSourcegraphButtonClassProps,
    onConfigureSourcegraphClick,
}: {
    sourcegraphURL: string
    ensureRepoExists: ViewOnSourcegraphButtonProps['ensureRepoExists']
    onConfigureSourcegraphClick?: ViewOnSourcegraphButtonProps['onConfigureSourcegraphClick']
} & Required<Pick<CodeHost, 'getContext'>> &
    Pick<CodeHost, 'viewOnSourcegraphButtonClassProps'>) => (mount: HTMLElement): void => {
    render(
        <ViewOnSourcegraphButton
            {...viewOnSourcegraphButtonClassProps}
            context={getContext()}
            sourcegraphURL={sourcegraphURL}
            ensureRepoExists={ensureRepoExists}
            onConfigureSourcegraphClick={onConfigureSourcegraphClick}
        />,
        mount
    )
}
