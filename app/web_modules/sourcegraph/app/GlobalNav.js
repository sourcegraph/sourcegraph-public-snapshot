// @flow

import React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import {Avatar, Button, Popover, Menu} from "sourcegraph/components";
import LogoutLink from "sourcegraph/user/LogoutLink";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import {LoginForm} from "sourcegraph/user/Login";
import {SignupForm} from "sourcegraph/user/Signup";

function GlobalNav({navContext, location}, {user, siteConfig, signedIn, router, eventLogger}) {
	return (
		<nav styleName="navbar" role="navigation">
			<Link to="/">
				<img styleName="logo" src={`${siteConfig.assetsRoot}/img/sourcegraph-mark.svg`}></img>
			</Link>
			<div styleName="context-container">{navContext}</div>

			<div styleName="actions">
				{user &&
					<div styleName="action-username">
						<Popover left={true}>
							{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <span>{user.Login}</span>}
							<Menu>
								<Link to="/">Your repositories</Link>
								<LogoutLink outline={true} size="small" block={true} />
							</Menu>
						</Popover>
					</div>
				}
				{!signedIn &&
					<div styleName="action">
						<LocationStateToggleLink href="/login" modalName="login" location={location}
							onToggle={(v) => v && eventLogger.logEvent("ShowLoginModal")}>
							<Button color="default" size="nav" block={true}>Sign in</Button>
						</LocationStateToggleLink>
					</div>
				}
				{location.state && location.state.modal === "login" &&
					<LocationStateModal modalName="login" location={location}
						onDismiss={(v) => eventLogger.logEvent("DismissLoginModal")}>
						<div styleName="modal">
							<LoginForm
								onLoginSuccess={dismissModal("login", location, router)}
								location={location} />
						</div>
					</LocationStateModal>
				}
				{!signedIn &&
					<div styleName="action">
						<LocationStateToggleLink href="/join" modalName="signup" location={location}
							onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal")}>
							<Button color="primary" size="nav" block={true}>Sign up</Button>
						</LocationStateToggleLink>
					</div>
				}
				{location.state && location.state.modal === "signup" &&
					<LocationStateModal modalName="signup" location={location}
						onDismiss={(v) => eventLogger.logEvent("DismissSignupModal")}>
						<div styleName="modal">
							<SignupForm
								onSignupSuccess={dismissModal("signup", location, router)}
								location={location} />
						</div>
					</LocationStateModal>
				}
			</div>
		</nav>
	);
}
GlobalNav.propTypes = {
	navContext: React.PropTypes.element,
	location: React.PropTypes.object.isRequired,
};
GlobalNav.contextTypes = {
	siteConfig: React.PropTypes.object.isRequired,
	user: React.PropTypes.object,
	signedIn: React.PropTypes.bool.isRequired,
	router: React.PropTypes.object.isRequired,
	eventLogger: React.PropTypes.object.isRequired,
};

export default CSSModules(GlobalNav, styles);
