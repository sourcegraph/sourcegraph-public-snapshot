import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { HelpPopover } from '../components/HelpPopover'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { OpenHelpPopoverButton } from '../global/OpenHelpPopoverButton'
import { eventLogger } from '../tracking/eventLogger'
import { UserAvatar } from '../user/UserAvatar'
import { canListAllRepositories, showDotComMarketing } from '../util/features'

interface Props {
    location: H.Location
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    adjacentToQueryInput?: boolean
    className?: string
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
}

const isGQLUser = (val: any): val is GQL.IUser => val && typeof val === 'object' && val.__typename === 'User'

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
            <div className={`nav-links ${this.props.className || ''}`}>
                {showDotComMarketing && (
                    <a
                        href="https://about.sourcegraph.com"
                        className="nav-links__border-link nav-links__ad"
                        onClick={this.onClickInstall}
                        title="Install self-hosted Sourcegraph to search your own code"
                    >
                        Install <span className="nav-links__widescreen-only">Sourcegraph</span>
                    </a>
                )}
                {this.props.user && (
                    <Link to="/search/searches" className="nav-links__link">
                        Searches
                    </Link>
                )}
                {canListAllRepositories && (
                    <Link to="/explore" className="nav-links__link">
                        Explore
                    </Link>
                )}
                {this.props.user &&
                    this.props.user.siteAdmin && (
                        <Link to="/site-admin" className="nav-links__link">
                            Admin
                        </Link>
                    )}
                {this.props.user && (
                    <Link className="nav-links__link nav-links__link-user" to="/settings/profile">
                        {isGQLUser(this.props.user) && this.props.user.avatarURL ? (
                            <UserAvatar size={64} />
                        ) : isGQLUser(this.props.user) ? (
                            this.props.user.username
                        ) : (
                            'Profile'
                        )}
                    </Link>
                )}
                <OpenHelpPopoverButton className="nav-links__help" onHelpPopoverToggle={this.onHelpPopoverToggle} />
                {this.props.showHelpPopover && <HelpPopover onDismiss={this.onHelpPopoverToggle} />}
                {!this.props.user &&
                    this.props.location.pathname !== '/sign-in' && (
                        <Link className="nav-links__link btn btn-primary" to="/sign-in">
                            Sign in
                        </Link>
                    )}
                <ThemeSwitcher {...this.props} className="nav-links__theme-switcher" />
                {showDotComMarketing && (
                    <a href="https://about.sourcegraph.com" className="nav-links__link">
                        About
                    </a>
                )}
            </div>
        )
    }

    private onHelpPopoverToggle = (): void => {
        this.props.onHelpPopoverToggle()
    }
}
