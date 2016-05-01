import React from "react";
import styles from "./styles/DashboardModal.css";
import CSSModules from "react-css-modules";
import Component from "sourcegraph/Component";
import Modal from "sourcegraph/components/Modal";


class DashboardModal extends Component {
	static propTypes = {
		header: React.PropTypes.string.isRequired,
		subheader: React.PropTypes.string.isRequired,
		body: React.PropTypes.string.isRequired,
		hasNext: React.PropTypes.bool.isRequired,
		onClick: React.PropTypes.func.isRequired,
		primaryCTA: React.PropTypes.func,
		secondaryCTA: React.PropTypes.func,
		img: React.PropTypes.func,
	}

	constructor(props) {
		super(props);
		this.state = {
			visible: true,
		};

		this.hide = this.hide.bind(this);
		this.show = this.show.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	show() {
		this.setState({visible: true});
	}

	hide() {
		if (!this.state.hasNext) {
			this.setState({visible: false});
		}
	}

	render() {
		if (!this.state.visible) return null;
		return (
			<Modal onDismiss={this.hide}>
				<div styleName="container">
					<div styleName="modal">
						<div styleName="header">
						<p styleName="header-text">{this.state.header}</p>
						</div>
						<div styleName="subheader">{this.state.subheader}</div>
						<div styleName="body">{this.state.body}</div>

						{this.state.img && this.state.img()}

						{this.state.primaryCTA && this.state.primaryCTA()}

						{this.state.secondaryCTA && this.state.secondaryCTA()}
					</div>
				</div>
			</Modal>
		);
	}
}

export default CSSModules(DashboardModal, styles);
