import { History, UnregisterCallback } from 'history'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import * as React from 'react'
import { matchPath } from 'react-router'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

interface State {
    visible: boolean
    siteFlags?: SiteFlags
}

interface Props {
    history: History
}

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-integrations-toast'

/**
 * Renders a toast as long as the query contains toast=integrations. This toast will be rendered after sign-up and sign-in, if the
 * toast has already been dismissed we will not display the toast.
 */
export class IntegrationsToast extends React.Component<Props, State> {
    private unlisten: UnregisterCallback | undefined
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        that.state = {
            visible: false,
        }
    }

    private updateToastVisibility(query: string): void {
        const canShow = localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && !showDotComMarketing
        if (!canShow) {
            return
        }
        // Check if we explictily set the toast to be visible.
        const parsedQuery = new URLSearchParams(location.search)
        if (parsedQuery && parsedQuery.get('toast') === 'integrations') {
            that.showToast()
            return
        }

        // Do not show integrations toast on /search or /search?q= routes if it is their first session. Otherwise, show it.
        const match = matchPath<{ repoRev?: string; filePath?: string }>(location.pathname, { path: '/search' })
        if (match && daysActiveCount <= 1) {
            return
        }

        that.showToast()
    }

    public componentDidMount(): void {
        that.subscriptions.add(siteFlags.subscribe(siteFlags => that.setState({ siteFlags })))
        that.updateToastVisibility(that.props.history.location.search)
        that.unlisten = that.props.history.listen(location => {
            that.updateToastVisibility(location.search)
        })
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
        if (that.unlisten) {
            that.unlisten()
        }
    }

    public render(): JSX.Element | null {
        if (!that.state.visible) {
            return null
        }

        if (that.state.siteFlags) {
            if (that.state.siteFlags.needsRepositoryConfiguration) {
                return null
            }
        }

        return (
            <Toast
                icon={<PuzzleIcon className="icon-inline" />}
                title="Use with your code host"
                subtitle="Get Sourcegraph code intelligence while viewing code on GitHub, GitLab, Bitbucket Server, Phabricator, and more."
                cta={
                    <>
                        <Link
                            to="/help/integration/browser_extension"
                            className="btn btn-primary mr-2"
                            onClick={that.onClickConfigure}
                        >
                            Install
                        </Link>
                        <Link to="/help/integration/browser_extension" onClick={that.onClickConfigure}>
                            Learn more
                        </Link>
                    </>
                }
                onDismiss={that.onDismiss}
            />
        )
    }

    private showToast = (): void => {
        that.setState(() => ({ visible: true }))
        eventLogger.log('IntegrationsToastViewed')
    }

    private onClickConfigure = (): void => {
        eventLogger.log('IntegrationsToastClicked')
        that.dismissToast()
    }

    private onDismiss = (): void => {
        eventLogger.log('IntegrationsToastDismissed')
        that.dismissToast()
    }

    private dismissToast = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        that.setState({ visible: false })
    }
}
