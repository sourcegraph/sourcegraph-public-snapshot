import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/dropdown.css";
import base from "./styles/_base.css";
import {CheckIcon} from "sourcegraph/components/Icons";
import {DownPointer} from "sourcegraph/components/symbols";

// This component is a minimal Dropdown component with some code copied from
// RevSwitcher.
class Dropdown extends React.Component {
	static propTypes = {
		icon: React.PropTypes.element,
		title: React.PropTypes.string.isRequired,
		initialValue: React.PropTypes.string,
		alwaysOpenMenu: React.PropTypes.bool, // Use initialValue to judge when false
		disabled: React.PropTypes.bool,
		items: React.PropTypes.arrayOf(React.PropTypes.shape({
			name: React.PropTypes.string,
			value: React.PropTypes.string,
		})).isRequired,
		onMenuClick: React.PropTypes.func,
		onItemClick: React.PropTypes.func,
		className: React.PropTypes.string,
	};

	constructor(props) {
		super(props);
		this.state = {
			open: false,
			selectedValue: this.props.initialValue,
		};
		this._onToggleDropdown = this._onToggleDropdown.bind(this);
		this._onDocumentClick = this._onDocumentClick.bind(this);
		this.getItemClickCallback = this.getItemClickCallback.bind(this);
	}

	componentDidMount() {
		if (typeof document !== "undefined") {
			document.addEventListener("click", this._onDocumentClick);
		}
	}

	componentWillUnmount() {
		if (typeof document !== "undefined") {
			document.removeEventListener("click", this._onDocumentClick);
		}
	}

	_onToggleDropdown(ev) {
		if (this.props.disabled) return;
		ev.preventDefault();
		ev.stopPropagation();
		this.setState({open: !this.state.open});
	}

	// _onDocumentClick causes clicks outside the menu to close the menu.
	_onDocumentClick(ev) {
		if (!this.state.open) return;
		if (this._wrapper && !this._wrapper.contains(ev.target)) this.setState({open: false});
	}

	getMenuClickCallback(val) {
		if (this.props.disabled) return () => null;
		if (this.props.alwaysOpenMenu || !this.props.initialValue) {
			return () => this.setState({open: !this.state.open});
		}
		return () => this.props.onMenuClick(val);
	}

	getItemClickCallback(val) {
		return () => {
			this.setState({selectedValue: val, open: false}, () => this.props.onItemClick(val));
		};
	}

	render() {
		return (
			<div styleName={this.props.disabled ? "wrapper_disabled" : "wrapper"} className={this.props.className}
				ref={(e) => this._wrapper = e}>
				<span onClick={this.getMenuClickCallback(this.state.selectedValue)}>{this.props.icon} {this.props.title}</span>
				<span className={`toggle ${this.state.open ? "open_arrow" : ""}`} onClick={this._onToggleDropdown}>
					<DownPointer width={8} styleName="center_icon" className={base.ml1}/>
				</span>
				<div styleName="dropdown_menu">
					<div role="menu"
						styleName={this.state.open ? "dropdown_menu_open" : "dropdown_menu_closed"}>
						<ul styleName="list_section">
							{this.props.items.map((item, i) =>
								<li key={i} styleName="item" onClick={this.getItemClickCallback(item.value)}>
									<span styleName="item_content">
										<CheckIcon styleName={item.value === this.state.selectedValue ? "item_icon" : "item_icon_hidden"} />
										<span styleName="item_name">{item.name}</span>
									</span>
								</li>
							)}
						</ul>
					</div>
				</div>
			</div>
		);
	}
}

export default CSSModules(Dropdown, styles);
