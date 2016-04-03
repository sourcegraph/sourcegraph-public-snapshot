// @flow

import React from "react";
import {Link} from "react-router";

import Dispatcher from "sourcegraph/Dispatcher";

import "sourcegraph/user/UserBackend"; // for side effects
import * as UserActions from "sourcegraph/user/UserActions";

import Style from "./styles/GlobalNav.css";
import {Avatar, Popover, Button} from "sourcegraph/components";
import context from "sourcegraph/context";


type Props = {
	navContext: Array<any>;
}

export default class GlobalNav extends React.Component {
	static propTypes = {
		navContext: React.PropTypes.element,
	};
	static defaultProps: {};
	props: Props;

	render() {
		// TODO: fix links
		const user = context.currentUser;
		return (
			<nav className={Style.navbar} role="navigation">
				<Link to="/">
					<img className={Style.logo} src={`${context.assetsRoot}/img/sourcegraph-mark.svg`}></img>
				</Link>

				<div className={Style.contextContainer}>{this.props.navContext}</div>

				<div className={Style.actions}>
					{user &&
						<div className={Style.action}>
							<div className={Style.userName}>
								<Popover left={true}>
									{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <span>{user.Login}</span>}
									<Button outline={true}
										small={true}
										block={true}
										onClick={() => Dispatcher.Backends.dispatch(new UserActions.SubmitLogout())}>Sign Out</Button>
								</Popover>
							</div>
						</div>
					}
					{!user &&
						<div className={Style.action}>
							<Link to="/join" className="sign-up">
								Sign up
							</Link>
						</div>
					}
					{!user &&
						<div className={Style.action}>
							<Link to="/login" className="sign-in">
								Sign in
							</Link>
						</div>
					}
				</div>
			</nav>
		);
	}
}
