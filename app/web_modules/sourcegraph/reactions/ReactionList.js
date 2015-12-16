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
		this.state.onSelect(reaction);
		// TODO A regular call to setState here does not work. React buffers calls to setState
		this.setState({}, () => this.setState({menu: null}));
	}

	render() {
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
	// TODO
};

export default ReactionList;
