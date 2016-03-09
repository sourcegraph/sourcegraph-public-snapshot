import React from "react";

import {annotate} from "sourcegraph/blob/Annotations";
import classNames from "classnames";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";

class BlobLine extends Component {
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
					if (ann.URLs) ann.URLs.forEach((url) => state.ownAnnURLs[url] = true);
				});
			}
		}

		// Filter highlightedDef to improve perf.
		state.highlightedDef = state.ownAnnURLs[props.highlightedDef] ? props.highlightedDef : null;
		state.activeDef = state.ownAnnURLs[props.activeDef] ? props.activeDef : null;

		state.lineNumber = props.lineNumber || null;
		state.oldLineNumber = props.oldLineNumber || null;
		state.newLineNumber = props.newLineNumber || null;
		state.startByte = props.startByte;
		state.contents = props.contents;
		state.selected = Boolean(props.selected);
		state.className = props.className || "";
		state.directLinks = Boolean(props.directLinks);
	}

	render() {
		const hasURL = (ann, url) => url && (ann.URL ? ann.URL === url : ann.URLs.includes(url));
		let i = 0;
		let contents = annotate(this.state.contents, this.state.startByte, this.state.annotations || [], (ann, content) => {
			i++;
			if (ann.URL || ann.URLs) {
				return (
					<a
						className={classNames(ann.Class, {
							"ref": true,
							"highlight-primary": hasURL(ann, this.state.highlightedDef),
							"active-def": hasURL(ann, this.state.activeDef),
						})}
						href={ann.URL || ann.URLs[0]}
						onMouseOver={() => Dispatcher.dispatch(new DefActions.HighlightDef(ann.URL || ann.URLs[0]))}
						onMouseOut={() => Dispatcher.dispatch(new DefActions.HighlightDef(null))}
						onClick={(ev) => {
							if (ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey || this.state.directLinks) return;
							ev.preventDefault();
							if (ann.URLs) {
								// Multiple refs coincident on the same token to different defs.
								//
								// Dispatch async and stop propagation so the menu is not
								// immediately closed by click handler on Document.
								Dispatcher.asyncDispatch(new DefActions.SelectMultipleDefs(
									ann.URLs,
									ev.view.scrollX + ev.clientX, ev.view.scrollY + ev.clientY
								));
							} else {
								Dispatcher.dispatch(new DefActions.SelectDef(ann.URL));
							}
						}}
						key={i}>{content}</a>
				);
			}
			return <span key={i} className={ann.Class}>{content.join("")}</span>;
		});

		let isDiff = this.state.oldLineNumber || this.state.newLineNumber;

		return (
			<tr className={`line ${this.state.selected ? "main-byte-range" : ""} ${this.state.className}`}
				data-line={this.state.lineNumber}>
				{this.state.lineNumber &&
					<td className="line-number"
						data-line={this.state.lineNumber}
						onClick={(event) => {
							if (event.shiftKey) {
								Dispatcher.dispatch(new BlobActions.SelectLineRange(this.state.lineNumber));
								return;
							}
							Dispatcher.dispatch(new BlobActions.SelectLine(this.state.lineNumber));
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

BlobLine.propTypes = {
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
	// destination instead of using pushState.
	directLinks: React.PropTypes.bool,
};

export default BlobLine;
