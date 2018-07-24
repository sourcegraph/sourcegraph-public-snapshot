import { ContributableMenu } from 'cxp/lib/protocol'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { HelpPopover } from '../components/HelpPopover'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { CXPCommandListPopoverButton } from '../cxp/components/CXPCommandList'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { ContributedActionsNavItems } from '../extensions/ContributedActions'
import { OpenHelpPopoverButton } from '../global/OpenHelpPopoverButton'
import { eventLogger } from '../tracking/eventLogger'
import { platformEnabled } from '../user/tags'
import { UserAvatar } from '../user/UserAvatar'
import { canListAllRepositories, showDotComMarketing } from '../util/features'

interface Props extends ExtensionsProps, ExtensionsChangeProps, CXPControllerProps {
    location: H.Location
    history: H.History
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
}

export class NavLinks extends React.PureComponent<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onClickInstall = (): void => {
        eventLogger.log('InstallSourcegraphServerCTAClicked', {
            location_on_page: 'Navbar',
        })
    }

    public render(): JSX.Element | null {
        return (
            <ul className="nav-links nav align-items-center pl-2 pr-1">
                {showDotComMarketing && (
                    <li className="nav-item">
                        <a
                            href="https://about.sourcegraph.com"
                            className="nav-link text-truncate"
                            onClick={this.onClickInstall}
                            title="Install self-hosted Sourcegraph to search your own code"
                        >
                            Install <span className="nav-links__widescreen-only">Sourcegraph</span>
                        </a>
                    </li>
                )}
                {platformEnabled(this.props.user) && (
                    <ContributedActionsNavItems
                        menu={ContributableMenu.GlobalNav}
                        cxpController={this.props.cxpController}
                    />
                )}
                {this.props.user && (
                    <li className="nav-item">
                        <Link to="/search/searches" className="nav-link">
                            Searches
                        </Link>
                    </li>
                )}
                {canListAllRepositories && (
                    <li className="nav-item">
                        <Link to="/explore" className="nav-link">
                            Explore
                        </Link>
                    </li>
                )}
                {this.props.user &&
                    this.props.user.siteAdmin && (
                        <li className="nav-item">
                            <Link to="/site-admin" className="nav-link">
                                Admin
                            </Link>
                        </li>
                    )}
                {this.props.user && (
                    <li className="nav-item">
                        <Link className="nav-link py-0 pr-2" to={`${this.props.user.url}/account`}>
                            {this.props.user.avatarURL ? (
                                <UserAvatar size={64} />
                            ) : (
                                <strong>{this.props.user.username}</strong>
                            )}
                        </Link>
                    </li>
                )}
                {platformEnabled(this.props.user) && (
                    <CXPCommandListPopoverButton
                        menu={ContributableMenu.CommandPalette}
                        cxpController={this.props.cxpController}
                    />
                )}
                <li className="nav-item">
                    <OpenHelpPopoverButton className="nav-link px-0" onHelpPopoverToggle={this.onHelpPopoverToggle} />
                </li>
                {this.props.showHelpPopover && (
                    <HelpPopover onDismiss={this.onHelpPopoverToggle} cxpController={this.props.cxpController} />
                )}
                {!this.props.user &&
                    this.props.location.pathname !== '/sign-in' && (
                        <li className="nav-item">
                            <Link className="nav-link btn btn-primary" to="/sign-in">
                                Sign in
                            </Link>
                        </li>
                    )}
                <li className="nav-item">
                    <ThemeSwitcher {...this.props} className="nav-link px-0" />
                </li>
                {showDotComMarketing && (
                    <li className="nav-item">
                        <a href="https://about.sourcegraph.com" className="nav-link">
                            About
                        </a>
                    </li>
                )}
            </ul>
        )
    }

    private onHelpPopoverToggle = (): void => {
        this.props.onHelpPopoverToggle()
    }
}
