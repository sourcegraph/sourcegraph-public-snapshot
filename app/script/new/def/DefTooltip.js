import React from "react";

import Component from "../Component";

class DefTooltip extends Component {
	constructor(props) {
		super(props);
		this._updatePosition = this._updatePosition.bind(this);
	}

	componentDidMount() {
		document.addEventListener("mousemove", this._updatePosition);
	}

	componentWillUnmount() {
		document.removeEventListener("mousemove", this._updatePosition);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_updatePosition(event) {
		this.setState({
			top: event.clientY + 15,
			left: Math.min(event.clientX + 15, window.innerWidth - 380),
		});
	}

	render() {
		let def = this.state.def;
		return (
			<div className="token-popover" style={{left: this.state.left, top: this.state.top}}>
				<div className="popover-data">
					<div className="title"><tt dangerouslySetInnerHTML={def.QualifiedName}></tt></div>
					<div className="content">
						<div className="doc" style={{maxHeight: 100, overflowY: "scroll"}} dangerouslySetInnerHTML={def.Data && def.Data.DocHTML}></div>
						<span className="repo">{def.Data.Repo}</span>
					</div>
				</div>
			</div>
		);
	}
}

DefTooltip.propTypes = {
	def: React.PropTypes.object,
};

export default DefTooltip;
