// @flow

import React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal} from "sourcegraph/components/Modal";
import {Avatar, Popover} from "sourcegraph/components";
import LogoutButton from "sourcegraph/user/LogoutButton";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import {LoginForm} from "sourcegraph/user/Login";
import {SignupForm} from "sourcegraph/user/Signup";

// dismissModal creates a function that dismisses the modal by setting
// the location state's stateKey property to false.
function dismissModal(stateKey, location, router) {
	return () => {
		router.replace({...location, state: {...location.state, [stateKey]: false}});
	};
}

function GlobalNav({navContext, location}, {user, siteConfig, signedIn, router, eventLogger}) {
	return (
		<nav styleName="navbar" role="navigation">
			<Link to="/">
				<img styleName="logo" src={`${siteConfig.assetsRoot}/img/sourcegraph-mark.svg`}></img>
			</Link>
			<div styleName="context-container">{navContext}</div>

			<div styleName="actions">
				{user &&
					<div styleName="action">
						<div styleName="username">
							<Popover left={true}>
								{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <span>{user.Login}</span>}
								<LogoutButton outline={true} size="small" block={true} />
							</Popover>
						</div>
					</div>
				}
				{!signedIn &&
					<div styleName="action">
						<LocationStateToggleLink href="/join" stateKey="signupModal" location={location}
							onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal")}>
							Sign up
						</LocationStateToggleLink>
						{location.state && location.state.signupModal &&
							<LocationStateModal stateKey="signupModal" location={location}
								onDismiss={(v) => eventLogger.logEvent("DismissSignupModal")}>
								<div styleName="modal">
									<SignupForm
										onSignupSuccess={dismissModal("signupModal", location, router)}
										location={location} />
								</div>
							</LocationStateModal>
						}
					</div>
				}
				{!signedIn &&
					<div styleName="action">
						<LocationStateToggleLink href="/login" stateKey="loginModal" location={location}
							onToggle={(v) => v && eventLogger.logEvent("ShowLoginModal")}>
							Sign in
						</LocationStateToggleLink>
						{location.state && location.state.loginModal &&
							<LocationStateModal stateKey="loginModal" location={location}
								onDismiss={(v) => eventLogger.logEvent("DismissLoginModal")}>
								<div styleName="modal">
									<LoginForm
										onLoginSuccess={dismissModal("loginModal", location, router)}
										location={location} />
								</div>
							</LocationStateModal>
						}
					</div>
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
