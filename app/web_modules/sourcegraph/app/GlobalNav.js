// @flow

import React from "react";
import {Link} from "react-router";

import Dispatcher from "sourcegraph/Dispatcher";

import "sourcegraph/user/UserBackend"; // for side effects
import * as UserActions from "sourcegraph/user/UserActions";

import {Avatar, Popover, Button} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";

function GlobalNav({navContext}, {user, siteConfig, signedIn}) {
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
								<Button outline={true}
									size="small"
									block={true}
									onClick={() => Dispatcher.Backends.dispatch(new UserActions.SubmitLogout())}>Sign Out</Button>
							</Popover>
						</div>
					</div>
				}
				{!signedIn &&
					<div styleName="action">
						<Link to="/join">
							Sign up
						</Link>
					</div>
				}
				{!signedIn &&
					<div styleName="action">
						<Link to="/login">
							Sign in
						</Link>
					</div>
				}
			</div>
		</nav>
	);
}
GlobalNav.propTypes = {
	navContext: React.PropTypes.element,
};
GlobalNav.contextTypes = {
	siteConfig: React.PropTypes.object.isRequired,
	user: React.PropTypes.object,
	signedIn: React.PropTypes.bool.isRequired,
};

export default CSSModules(GlobalNav, styles);
