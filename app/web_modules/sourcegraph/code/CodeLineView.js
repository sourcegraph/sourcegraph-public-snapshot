import React from "react";

import classNames from "classnames";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as CodeActions from "sourcegraph/code/CodeActions";
import * as DefActions from "sourcegraph/def/DefActions";

class CodeLineView extends Component {
	constructor(props) {
		super(props);
		this.state = {
			ownAnnURLs: {},
		};
	}

	reconcileState(state, props) {
		// Update ownAnnURLs when they change.
		if (state.annotations !== props.annotations) {
			state.annotations = props.annotations;
			state.ownAnnURLs = {};
			if (state.annotations) {
				state.annotations.forEach((ann) => {
					if (ann.URL) state.ownAnnURLs[ann.URL] = true;
				});
			}
		}

		// Filter highlightedDef to improve perf.
		state.highlightedDef = state.ownAnnURLs[props.highlightedDef] ? props.highlightedDef : null;
		state.activeDef = state.ownAnnURLs[props.activeDef] ? props.activeDef : null;

		state.lineNumber = props.lineNumber || null;
		state.oldLineNumber = props.oldLineNumber || null;
		state.newLineNumber = props.newLineNumber || null;
		state.startByte = props.startByte || null;
		state.contents = props.contents;
		state.selected = Boolean(props.selected);
		state.className = props.className || "";
		state.directLinks = Boolean(props.directLinks);
	}

	render() {
		let contents;
		if (this.state.annotations) {
			contents = [];
			let pos = 0;
			let skip;
			this.state.annotations.forEach((ann, i) => {
				if (skip >= i) {
					// This annotation's class was already merged into a previous annotation.
					return;
				}

				let cls;
				let extraURLs;

				// Merge syntax highlighting and multiple-def annotations into the previous link, if any.
				for (let j = i + 1; j < this.state.annotations.length; j++) {
					let ann2 = this.state.annotations[j];
					if (ann2.StartByte === ann.StartByte && ann2.EndByte === ann.EndByte) {
						if (ann2.Class) {
							cls = cls || [];
							cls.push(ann2.Class);
						}
						if (ann2.URL) {
							extraURLs = extraURLs || [];
							extraURLs.push(ann2.URL);
						}
						skip = j;
					} else {
						break;
					}
				}

				const start = ann.StartByte - this.state.startByte;
				const end = ann.EndByte - this.state.startByte;
				if (start > pos) {
					contents.push(this.state.contents.slice(pos, start));
				}

				let matchesURL = (url) => ann.URL === url || (extraURLs && extraURLs.includes(url));

				const text = this.state.contents.slice(start, end);
				let el;
				if (ann.URL) {
					el = (
						<a
							className={classNames(cls, {
								"ref": true,
								"highlight-primary": matchesURL(this.state.highlightedDef),
								"active-def": matchesURL(this.state.activeDef),
							})}
							href={ann.URL}
							onMouseOver={() => Dispatcher.dispatch(new DefActions.HighlightDef(ann.URL))}
							onMouseOut={() => Dispatcher.dispatch(new DefActions.HighlightDef(null))}
							onClick={(ev) => {
								if (ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey || this.state.directLinks) return;
								ev.preventDefault();
								if (extraURLs) {
									Dispatcher.asyncDispatch(new DefActions.SelectMultipleDefs([ann.URL].concat(extraURLs), ev.view.scrollX + ev.clientX, ev.view.scrollY + ev.clientY)); // dispatch async so that the menu is not immediately closed by click handler on document
								} else {
									Dispatcher.asyncDispatch(new DefActions.SelectDef(ann.URL));
								}
							}}
							key={i}>{text}</a>
					);
				} else {
					el = <span key={i} className={ann.Class}>{text}</span>;
				}
				contents.push(el);
				pos = end;
			});
			if (pos < this.state.contents.length) {
				contents.push(this.state.contents.slice(pos));
			}
		} else {
			contents = this.state.contents;
		}

		let isDiff = this.state.oldLineNumber || this.state.newLineNumber;

		return (
			<tr className={`line ${this.state.selected ? "main-byte-range" : ""} ${this.state.className}`}
				data-line={this.state.lineNumber}>
				{this.state.lineNumber &&
					<td className="line-number"
						data-line={this.state.lineNumber}
						onClick={(event) => {
							if (event.shiftKey) {
								Dispatcher.dispatch(new CodeActions.SelectLineRange(this.state.lineNumber));
								return;
							}
							Dispatcher.dispatch(new CodeActions.SelectLine(this.state.lineNumber));
						}}>
					</td>}
				{isDiff && <td className="line-number" data-line={this.state.oldLineNumber || ""}></td>}
				{isDiff && <td className="line-number" data-line={this.state.newLineNumber || ""}></td>}

				<td className="line-content">
					{contents}
					{this.state.contents === "" && <span>&nbsp;</span>}
				</td>
			</tr>
		);
	}
}

CodeLineView.propTypes = {
	lineNumber: (props, propName, componentName) => {
		let v = React.PropTypes.number(props, propName, componentName);
		if (v) return v;
		if (typeof props.lineNumber !== "undefined" && (typeof props.oldLineNumber !== "undefined" || typeof props.newLineNumber !== "undefined")) {
			return new Error("If lineNumber is set, then oldLineNumber/newLineNumber (which are for diff hunks) may not be used");
		}
	},

	// For diff hunks.
	oldLineNumber: React.PropTypes.number,
	newLineNumber: React.PropTypes.number,

	// startByte is the byte position of the first byte of contents. It is
	// required if annotations are specified, so that the annotations can
	// be aligned to the contents.
	startByte: (props, propName, componentName) => {
		if (props.annotations) return React.PropTypes.number.isRequired(props, propName, componentName);
	},
	contents: React.PropTypes.string,
	annotations: React.PropTypes.array,
	selected: React.PropTypes.bool,
	highlightedDef: React.PropTypes.string,
	className: React.PropTypes.string,

	// directLinks, if true, makes clicks on annotation links go directly to the
	// destination instead of using GoTo and pushState.
	directLinks: React.PropTypes.bool,
};

export default CodeLineView;
