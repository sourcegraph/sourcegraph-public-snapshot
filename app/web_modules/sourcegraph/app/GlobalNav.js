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
						<a href="https://text.sourcegraph.com" styleName="logged-out-nav-item">Blog</a>
						<Link to="/about" styleName="logged-out-nav-item">About</Link>
						<Link to="/pricing" styleName="logged-out-nav-item">Pricing</Link>
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
				<div styleName="flex-fill">
					{user && <div styleName="flex flex-end">
						{!navContext && <div styleName="flex-fill tl" className={base.bn}>
							<Link to="/tour">
								<TabItem active={location.pathname === "/tour"} icon="tour">Tour</TabItem>
							</Link>
							<Link to="/repositories">
								<TabItem active={location.pathname === "/repositories"} icon="repository">My Repositories</TabItem>
							</Link>
							<Link to="/tools">
								<TabItem hideMobile={true} active={location.pathname === "/tools"} icon="tools">Tools</TabItem>
							</Link>
							<Link to="/search">
								<TabItem active={location.pathname === "/search"} icon="search">
									<span styleName="hidden-s">Code</span> Search
								</TabItem>
							</Link>
						</div>}
						<div styleName="flex" className={`${base.pv2} ${base.ph3}`}>
						{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal styleName="icon-ellipsis" title="Your editor could not identify the symbol"/>}
							{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon styleName="icon-check" title="Sourcegraph successfully looked up symbol" />}
							<Popover left={true}>
								{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} styleName="block" className={base.pt2} /> : <div styleName="username">{user.Login}</div>}
								<Menu>
									<Link to="/">Home</Link>
									<LogoutLink outline={true} size="small" block={true} />
								</Menu>
							</Popover>
						</div>
					</div>}
					{!signedIn &&
						<div styleName="login-signup-links" className={base.pv2}>
							{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal styleName="icon-ellipsis" title="Your editor could not identify the symbol"/>}
							{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon styleName="icon-check" title="Sourcegraph successfully looked up symbol" />}
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
