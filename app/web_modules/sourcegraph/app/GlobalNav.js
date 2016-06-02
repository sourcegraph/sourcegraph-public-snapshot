// @flow

import React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import {Avatar, Popover, Menu, Button, TabItem, Logo} from "sourcegraph/components";
import LogoutLink from "sourcegraph/user/LogoutLink";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import base from "sourcegraph/components/styles/_base.css";
import {LoginForm} from "sourcegraph/user/Login";
import {SignupForm} from "sourcegraph/user/Signup";
import {EllipsisHorizontal, CheckIcon} from "sourcegraph/components/Icons";

function GlobalNav({navContext, location, channelStatusCode}, {user, siteConfig, signedIn, router, eventLogger}) {
	if (location.pathname === "/styleguide") return <span />;
	return (
		<nav styleName="navbar" role="navigation">

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
			<div styleName="flex flex-fill flex-center tl" className={base.bn}>
				<Link to="/">
					<Logo styleName={`logo flex-fixed ${signedIn ? "logomark" : ""}`}
						width={signedIn ? "24px" : "200px"}
						type={signedIn ? "logomark" : "logotype"}/>
				</Link>
				{user && <div styleName="flex flex-start flex-item-auto">
					<Link to="/tour">
						<TabItem active={location.pathname === "/tour"} icon="tour">Tour</TabItem>
					</Link>
					<Link to="/repositories">
						<TabItem active={location.pathname === "/repositories"} icon="repository">My Repositories</TabItem>
					</Link>
					<Link to="/tools">
						<TabItem hideMobile={true} active={location.pathname === "/tools"} icon="tools">Tools</TabItem>
					</Link>
					<Link to="/">
						<TabItem active={location.pathname === "/search"} icon="search">
							<span styleName="hidden-s">Code</span> Search
						</TabItem>
					</Link>
				</div>}
				{!user && <div styleName="flex-start flex-item-auto hidden-s">
					<Link to="/about" styleName="logged-out-nav-item">About</Link>
					<Link to="/pricing" styleName="logged-out-nav-item">Pricing</Link>
					<a href="https://text.sourcegraph.com" styleName="logged-out-nav-item">Blog</a>
				</div>}
				{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal styleName="icon-ellipsis" title="Your editor could not identify the symbol"/>}
				{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon styleName="icon-check" title="Sourcegraph successfully looked up symbol" />}

				{user && <div styleName="flex flex-fixed" className={`${base.pv2} ${base.ph3}`}>
					<Popover left={true}>
						{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} styleName="block" className={base.pt2} /> : <div styleName="username">{user.Login}</div>}
						<Menu>
							<Link to="/about" role="menu-item">About</Link>
							<Link to="/contact" role="menu-item">Contact</Link>
							<a href="https://boards.greenhouse.io/sourcegraph" target="_blank" role="menu-item">We're hiring</a>
							<Link to="/security" role="menu-item">Security</Link>
							<Link to="/-/privacy" role="menu-item">Privacy</Link>
							<Link to="/-/terms" role="menu-item">Terms</Link>
							<hr className={base.m0} />
							<LogoutLink role="menu-item" />
						</Menu>
					</Popover>
				</div>}

				{!signedIn &&
					<div styleName="tr" className={base.pv2}>
						<div styleName="action">
							<LocationStateToggleLink href="/login" modalName="login" location={location}
								onToggle={(v) => v && eventLogger.logEvent("ShowLoginModal")}>
								<Button color="blue" outline={true}>Sign in</Button>
							</LocationStateToggleLink>
						</div>
						<div styleName="action hidden-s">
							<LocationStateToggleLink href="/join" modalName="signup" location={location}
								onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal")}>
								<Button color="blue">Sign up</Button>
							</LocationStateToggleLink>
						</div>
					</div>
				}
			</div>
		</nav>
	);
}

GlobalNav.propTypes = {
	navContext: React.PropTypes.element,
	location: React.PropTypes.object.isRequired,
	channelStatusCode: React.PropTypes.number,
};
GlobalNav.contextTypes = {
	siteConfig: React.PropTypes.object.isRequired,
	user: React.PropTypes.object,
	signedIn: React.PropTypes.bool.isRequired,
	router: React.PropTypes.object.isRequired,
	eventLogger: React.PropTypes.object.isRequired,
};

export default CSSModules(GlobalNav, styles, {allowMultiple: true});
