import * as React from 'react'
import { Subscription } from 'rxjs'
import { browserExtensionInstalled } from '../tracking/analyticsUtils'
import { eventLogger } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const CHROME_EXTENSION_STORE_LINK = 'https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack'
const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-browser-ext-toast'

interface Props {
    browserLogoAsset: string
    browserName: string
    link: string
    onClickInstall: () => void
}

interface State {
    visible: boolean
}

abstract class BrowserExtensionToast extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            visible: false,
        }
    }

    public componentDidMount(): void {
        // Display if we don't receive confirmation that the user already has
        // the extension installed within a short time.
        this.subscriptions.add(
            browserExtensionInstalled.subscribe(isInstalled => {
                const visible =
                    !isInstalled &&
                    showDotComMarketing &&
                    localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' &&
                    daysActiveCount === 1
                this.setState({ visible })
                if (visible) {
                    eventLogger.log('BrowserExtReminderViewed')
                }
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<img className="logo-icon" src={this.props.browserLogoAsset} />}
                title="Get Sourcegraph on GitHub"
                subtitle={`Get code intelligence while browsing GitHub and reading PRs with the Sourcegraph ${this.props.browserName} extension`}
                cta={
                    <a
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn btn-primary"
                        href={this.props.link}
                        onClick={this.onClickInstall}
                    >
                        Install
                    </a>
                }
                onDismiss={this.onDismiss}
            />
        )
    }

    private onClickInstall = (): void => {
        this.props.onClickInstall()
        this.onDismiss()
    }

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}

export class ChromeExtensionToast extends React.Component {
    public render(): JSX.Element | null {
        return (
            <BrowserExtensionToast
                browserName="Chrome"
                browserLogoAsset="/.assets/img/logo-chrome.svg"
                onClickInstall={this.onClickInstall}
                link={CHROME_EXTENSION_STORE_LINK}
            />
        )
    }

    private onClickInstall = (): void =>
        eventLogger.log('BrowserExtInstallClicked', { marketing: { browser: 'Chrome' } })
}
