import React from "react";
import Component from "sourcegraph/Component";
import s from "sourcegraph/def/styles/Def.css";

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
		this._elem = null;
		document.removeEventListener("mousemove", this._updatePosition);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_updatePosition(event) {
		if (!this._elem) return;
		if (typeof window !== "undefined") {
			window.requestAnimationFrame(() => {
				if (!this._elem) return;
				this._elem.style.top = `${event.clientY + 15}px`;
				this._elem.style.left = `${Math.min(event.clientX + 15, window.innerWidth - 380)}px`;
			});
		}
	}

	render() {
		let def = this.state.def;

		let inner;
		if (def.Error) {
			inner = <span className={s.error}>Definition not available</span>;
		} else {
			inner = [
				<div key="title" className={s.tooltipTitle} dangerouslySetInnerHTML={def.QualifiedName}></div>,
				<div key="content" className={s.content}>
					{def && def.DocHTML && <div className={s.doc} dangerouslySetInnerHTML={def && def.DocHTML}></div>}
					{def && def.Repo !== this.state.currentRepo && <span className={s.repo}>{def.Repo}</span>}
				</div>,
			];
		}

		return (
			<div ref={(e) => this._elem = e} className={def.Error ? s.tooltipError : s.tooltipFound}>
				{inner}
			</div>
		);
	}
}

DefTooltip.propTypes = {
	// currentRepo is the repo of the file that's currently being displayed, if any.
	currentRepo: React.PropTypes.string,

	def: React.PropTypes.object,
};

export default DefTooltip;
