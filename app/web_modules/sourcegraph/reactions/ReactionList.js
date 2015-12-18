import React from "react";

import Component from "sourcegraph/Component";
import EmojiMenu from "sourcegraph/reactions/EmojiMenu";
import Tooltip from "sourcegraph/util/Tooltip";
import * as emoji from "sourcegraph/reactions/emoji";

import classNames from "classnames";

class ReactionList extends Component {
	constructor(props) {
		super(props);
		this.state = {
			menu: null,
		};
		this._showMenu = this._showMenu.bind(this);
		this._onSelect = this._onSelect.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_showMenu(event) {
		this.setState({menu: {x: event.clientX, y: event.clientY + 20}});
	}

	_onSelect(reaction) {
		// TODO A normal call to setState here does not work due to side-effects
		// of React's buffered updates to this.state and the behavior of
		// Component._updateState.
		this.setState({menu: null}, () => { this.state.onSelect(reaction); });
	}

	render() {
		if (this.state.reactions.length === 0) return null;

		return (
			<div className="reaction-list">
				<div className="reactions">
					{this.state.reactions.map((reaction) => {
						let classes = classNames({
							"reaction": true,
							"reaction-list-item": true,
							"user-reacted": reaction.Users.some((user) => user.Login === this.state.currentUser.Login),
						});
						let usernames = reaction.Users.map((user) => user.Login);
						if (usernames.length > 1) usernames[usernames.length-1] = `and ${usernames[usernames.length-1]}`;

						return (
							<div className={classes} key={reaction.Reaction} onClick={() => this._onSelect(reaction.Reaction)} style={{position: "relative"}}>
								<Tooltip>
									<b>{usernames.join(", ")}</b> reacted with <b>{reaction.Reaction}</b>
								</Tooltip>
								<img className="emoji" src={emoji.url(reaction.Reaction)} />
								<b>{reaction.Users.length}</b>
							</div>
						);
					})}
				</div>
				<div className="reaction-list-item add-reaction" onClick={this._showMenu} style={this.state.menu ? {opacity: 1} : null}>
					<i className="fa fa-smile-o"></i><sup>+</sup>
				</div>
				{this.state.menu &&
					<EmojiMenu x={this.state.menu.x} y={this.state.menu.y} onSelect={this._onSelect} onClose={() => this.setState({menu: null})} />
				}
			</div>
		);
	}
}

ReactionList.propTypes = {
	reactions: React.PropTypes.arrayOf(React.PropTypes.shape({
		Reaction: React.PropTypes.string,
		Users: React.PropTypes.arrayOf(React.PropTypes.shape({
			Login: React.PropTypes.string,
		})),
	})).isRequired,
	currentUser: React.PropTypes.shape({
		Login: React.PropTypes.string,
	}).isRequired,
	onSelect: React.PropTypes.func.isRequired,
};

export default ReactionList;
