import { ActionsNavItems } from '@sourcegraph/extensions-client-common/lib/app/actions/ActionsNavItems'
import { CommandListPopoverButton } from '@sourcegraph/extensions-client-common/lib/app/CommandList'
import { CXPControllerProps } from '@sourcegraph/extensions-client-common/lib/cxp/controller'
import { ContributableMenu } from 'cxp/module/protocol'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { HelpPopover } from '../components/HelpPopover'
import { HistoryPopoverContainer } from '../components/HistoryPopoverContainer'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'
import { OpenHelpPopoverButton } from '../global/OpenHelpPopoverButton'
import { eventLogger } from '../tracking/eventLogger'
import { platformEnabled } from '../user/tags'
import { UserAvatar } from '../user/UserAvatar'
import { canListAllRepositories, showDotComMarketing } from '../util/features'

interface Props extends ExtensionsProps, CXPControllerProps {
    location: H.Location
    history: H.History
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
}

const fileHistoryEnabled = localStorage.getItem('enable-file-history') === 'true'

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
                    <ActionsNavItems
                        menu={ContributableMenu.GlobalNav}
                        cxpController={this.props.cxpController}
                        extensions={this.props.extensions}
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
                {platformEnabled(this.props.user) && (
                    <li className="nav-item">
                        <Link to="/extensions" className="nav-link">
                            Extensions
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
                {platformEnabled(this.props.user) && (
                    <CommandListPopoverButton
                        menu={ContributableMenu.CommandPalette}
                        cxpController={this.props.cxpController}
                        extensions={this.props.extensions}
                    />
                )}
                {fileHistoryEnabled && (
                    <li>
                        <HistoryPopoverContainer location={this.props.location} history={this.props.history} />
                    </li>
                )}
                {this.props.user ? (
                    <li className="nav-item">
                        <Link className="nav-link py-0" to={`${this.props.user.url}/account`}>
                            {this.props.user.avatarURL ? (
                                <UserAvatar size={64} />
                            ) : (
                                <strong>{this.props.user.username}</strong>
                            )}
                        </Link>
                    </li>
                ) : (
                    this.props.location.pathname !== '/sign-in' && (
                        <li className="nav-item mx-1">
                            <Link className="nav-link btn btn-primary" to="/sign-in">
                                Sign in
                            </Link>
                        </li>
                    )
                )}
                <li className="nav-item">
                    <OpenHelpPopoverButton className="nav-link px-0" onHelpPopoverToggle={this.onHelpPopoverToggle} />
                </li>
                {this.props.showHelpPopover && (
                    <HelpPopover
                        onDismiss={this.onHelpPopoverToggle}
                        cxpController={this.props.cxpController}
                        extensions={this.props.extensions}
                    />
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
