import React from "react";

import {CodeIcon, TextIcon} from "./Icons";

import CSSModules from "react-css-modules";
import styles from "./App.css";

@CSSModules(styles)
export default class SearchMenu extends React.Component {
	static propTypes = {
		onSelect: React.PropTypes.func,
		selected: React.PropTypes.string
	};

	constructor(props) {
		super(props);
		this.state = {
			selected: this.props.selected || "def",
		};
	}

	select = (selected) => {
		this.setState({selected}, () => {
			if (this.props.onSelect) {
				this.props.onSelect(selected);
			}
		});
	};

	render() {
		return (
			<nav className="menu">
				<div styleName={this.state.selected === "def" ? "menu-item-selected" : "menu-item"} onClick={() => this.select("def")}>
					<CodeIcon /><span style={{marginLeft: "10px"}}>Code</span>
				</div>
				<div styleName={this.state.selected === "text" ? "menu-item-selected" : "menu-item"} onClick={() => this.select("text")}>
					<TextIcon /><span style={{marginLeft: "10px"}}>Text</span>
				</div>
			</nav>
		);
	}
}
