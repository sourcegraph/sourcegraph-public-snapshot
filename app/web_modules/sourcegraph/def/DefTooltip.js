import React from "react";

import Component from "sourcegraph/Component";

// function created to update cursor position in constructor()
let cursorX;
let cursorY;
if (typeof document !== "undefined") {
	// TODO(autotest) support document object.
	document.addEventListener("mousemove", (event) => {
		cursorX = event.clientX;
		cursorY = event.clientY;
	}, false);
}

class DefTooltip extends Component {
	constructor(props) {
		super(props);
		this._updatePosition = this._updatePosition.bind(this);
		// TODO(autotest) support document object.
		if (typeof window !== "undefined") {
			this.state = {
				top: cursorY + 15,
				left: Math.min(cursorX + 15, window.innerWidth - 380),
			};
		} else {
			this.state = {
				top: cursorY + 15,
				left: Math.min(cursorX + 15, 0),
			};
		}
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
					<div className="title"><code dangerouslySetInnerHTML={def.QualifiedName}></code></div>
					<div className="content">
						<div className="doc" style={{maxHeight: 100, overflowY: "hidden"}} dangerouslySetInnerHTML={def.Data && def.Data.DocHTML}></div>
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
