import React from "react";
import ReactDOM from "react-dom";

import Component from "sourcegraph/Component";

import * as emoji from "sourcegraph/reactions/emoji";

class EmojiMenu extends Component {
	constructor(props) {
		super(props);
		this.state = {
			currentHover: null,
		};
		this._onKeyDown = this._onKeyDown.bind(this);
		this._onSelect = this._onSelect.bind(this);
		this._onClick = this._onClick.bind(this);
	}

	componentDidMount() {
		document.addEventListener("keydown", this._onKeyDown);
		document.addEventListener("mousedown", this._onClick);
	}

	componentWillUnmount() {
		document.removeEventListener("keydown", this._onKeyDown);
		document.removeEventListener("mousedown", this._onClick);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_close() {
		if (this.state.onClose) this.state.onClose();
	}

	_onKeyDown(event) {
		// Close via "esc" key.
		if (event.keyCode === 27) {
			this._close();
		}
	}

	_onClick(event) {
		let el = ReactDOM.findDOMNode(this);
		// Close when a click event happens outside of this element.
		if (!el.contains(event.target)) this._close();
	}

	_onSelect(name) {
		this.state.onSelect(name);
		this._close();
	}

	_onEmojiHover(name) {
		this.setState({currentHover: name});
	}

	render() {
		const width = 360;
		const listHeight = 260;
		const footerHeight = 64;
		let height = listHeight + footerHeight;
		let menuStyle = {
			width: width,
			left: Math.min(this.state.x, (window.innerWidth - width) - 10),
			top: Math.min(this.state.y, (window.innerHeight - height) - 10),
		};

		return (
			<div className="emoji-menu" style={menuStyle}>
				<div className="emoji-list" style={{height: listHeight}}>
					{emoji.list().map((name) => <img
						key={name}
						className="emoji"
						src={emoji.url(name)}
						onMouseOver={() => { this._onEmojiHover(name); }}
						onClick={() => { this._onSelect(name); }}/>)}
				</div>
				<div className="emoji-footer" style={{height: footerHeight}}>
					{this.state.currentHover &&
						<span>
							<img className="emoji" src={emoji.url(this.state.currentHover)} />
							<b>{this.state.currentHover}</b>
						</span>
					}
				</div>
			</div>
		);
	}
}

EmojiMenu.propTypes = {
	// TODO
};

export default EmojiMenu;
