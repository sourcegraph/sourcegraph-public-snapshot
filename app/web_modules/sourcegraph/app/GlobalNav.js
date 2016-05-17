// @flow

import React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import Logo from "sourcegraph/components/Logo";
import {Avatar, Popover, Menu, Button} from "sourcegraph/components";
import LogoutLink from "sourcegraph/user/LogoutLink";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import base from "sourcegraph/components/styles/base.css";
import {LoginForm} from "sourcegraph/user/Login";
import {SignupForm} from "sourcegraph/user/Signup";

function GlobalNav({navContext, location}, {user, siteConfig, signedIn, router, eventLogger}) {
	return (
		<nav styleName={signedIn || location.pathname !== "/" ? "navbar" : ""} role="navigation">

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

			{location.state && location.state.modal === "signup" &&
				<LocationStateModal modalName="signup" location={location}
					onDismiss={(v) => eventLogger.logEvent("DismissSignupModal")}>
					<div styleName="modal">
						<SignupForm
							onSignupSuccess={dismissModal("signup", location, router, {_onboarding: "new-user", _signupChannel: "email"})}
							location={location} />
					</div>
				</LocationStateModal>
			}

			{!signedIn && location.pathname === "/" &&
				// TODO(chexee): design a consistent header with a few different states. The code here should not require so much specific logic.
				<div styleName="logged-out-header">
					<Link to="/" styleName="header-logo"><Logo width="220px" type="logotype" /></Link>
					<nav styleName="logged-out-nav">
						<a href="/blog" styleName="logged-out-nav-item">Blog</a>
						<a href="/about" styleName="logged-out-nav-item">About</a>
						<LocationStateToggleLink href="/login"
							modalName="login"
							location={location}
							onToggle={(v) => v && eventLogger.logEvent("ShowLoginModal")}
							className={base.mh3}>
							<Button color="blue" outline={true}>Log in</Button>
						</LocationStateToggleLink>
						<LocationStateToggleLink href="/join"
							modalName="signup" location={location}
							onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal")}>
							<Button color="blue">Sign up</Button>
						</LocationStateToggleLink>
					</nav>
				</div>
			}

			{(signedIn || location.pathname !== "/") &&
				<Link to="/">
					<Logo styleName="logo" width="24px" />
				</Link>
			}

			{(signedIn || location.pathname !== "/") && <div styleName="context-container">{navContext}</div>}

			{(signedIn || location.pathname !== "/") &&
				<div styleName="actions">
					{user && <div style={{display: "inline-flex", alignItems: "center"}}>
						<Link styleName="action" to="/repositories">Repositories</Link>
						<Link styleName="action" to="/tools">Tools</Link>
						<div styleName="action-username">
							<Popover left={true}>
								{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <span>{user.Login}</span>}
								<Menu>
									<Link to="/">Home</Link>
									<LogoutLink outline={true} size="small" block={true} />
								</Menu>
							</Popover>
						</div>
					</div>}
					{!signedIn &&
						<div>
							<div styleName="action">
								<LocationStateToggleLink href="/login" modalName="login" location={location}
									onToggle={(v) => v && eventLogger.logEvent("ShowLoginModal")}>
									<Button color="blue" outline={true}>Sign in</Button>
								</LocationStateToggleLink>
							</div>
							<div styleName="action">
								<LocationStateToggleLink href="/join" modalName="signup" location={location}
									onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal")}>
									<Button color="blue">Sign up</Button>
								</LocationStateToggleLink>
							</div>
						</div>
					}
				</div>
			}
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
