import * as React from "react";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import s from "sourcegraph/def/styles/Def.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import type {BlobPos} from "sourcegraph/def/DefActions";
import * as DefActions from "sourcegraph/def/DefActions";

// These variables are needed to intialize the tooltips position to the current
// position of the mouse without a mousemove event.
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
	}

	componentDidMount() {
		document.addEventListener("mousemove", this._updatePosition);
	}

	componentWillUnmount() {
		this._elem = null;
		document.removeEventListener("mousemove", this._updatePosition);
	}

	props: {
		currentRepo: string,
		hoverPos: BlobPos,
		hoverInfos: any,
	};

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	onStateTransition(prevState, nextState) {
		if (prevState.hoverPos !== nextState.hoverPos && nextState.hoverPos !== null) {
			Dispatcher.Backends.dispatch(new DefActions.WantHoverInfo(nextState.hoverPos));
		}
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
		if (this.state.hoverPos === null) {
			return null;
		}

		let info = this.state.hoverInfos.get(this.state.hoverPos);
		if (info === null) {
			// TODO show loading indicator
			return null;
		}

		let def = info.def;
		return (
			<div ref={(e) => { this._elem = e; this._updatePosition({clientY: cursorY, clientX: cursorX}); }} className={s.tooltip}>
				<div key="title" className={s.tooltipTitle}>{info.Title || qualifiedNameAndType(def)}</div>
				<div key="content" className={s.content}>
					{def && def.DocHTML && <div className={s.doc} dangerouslySetInnerHTML={def && def.DocHTML}></div>}
					{def && def.Repo !== this.state.currentRepo && <span className={s.repo}>{def.Repo}</span>}
				</div>
			</div>
		);
	}
}

export default DefTooltip;
