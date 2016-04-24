// @flow

import React from "react";
import {Link} from "react-router";

import Dispatcher from "sourcegraph/Dispatcher";

import "sourcegraph/user/UserBackend"; // for side effects
import * as UserActions from "sourcegraph/user/UserActions";

import {Avatar, Popover, Button} from "sourcegraph/components";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";

type Props = {
	navContext: Array<any>;
};

class GlobalNav extends React.Component {
	static propTypes = {
		navContext: React.PropTypes.element,
	};
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};
	static defaultProps: {};
	props: Props;

	render() {
		const user = context.currentUser;
		return (
			<nav styleName="navbar" role="navigation">
				<Link to="/">
					<img styleName="logo" src={`${this.context.siteConfig.assetsRoot}/img/sourcegraph-mark.svg`}></img>
				</Link>
				<div styleName="context-container">{this.props.navContext}</div>

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
					{!user &&
						<div styleName="action">
							<Link to="/join">
								Sign up
							</Link>
						</div>
					}
					{!user &&
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
}

export default CSSModules(GlobalNav, styles);
