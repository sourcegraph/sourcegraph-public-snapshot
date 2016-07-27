import * as React from "react";
import {Link} from "react-router";
import utf8 from "utf8";

import {annotate} from "sourcegraph/blob/Annotations";
import classNames from "classnames";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import {urlToBlob} from "sourcegraph/blob/routes";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
import {fastURLToRepoDef} from "sourcegraph/def/routes";
import s from "sourcegraph/blob/styles/Blob.css";
import {isExternalLink} from "sourcegraph/util/externalLink";
import "sourcegraph/components/styles/code.css";

// simpleContentsString converts [string...] (like ["a", "b", "c"]) to
// a string by joining the elements (to produce "abc", for example).
function simpleContentsString(contents) {
	if (!(contents instanceof Array)) return contents;
	if (contents.some((e) => typeof e !== "string")) return contents;
	return contents.join("");
}

// converts each string component of contents from UTF-8
function fromUtf8(contents) {
	if (typeof contents === "string") {
		return utf8.decode(contents);
	}
	if (!(contents instanceof Array)) {
		return contents;
	}
	return contents.map((e) =>
		typeof e !== "string" ? e : utf8.decode(e)
	);
}

// fastInsertRevIntoDefURL accepts a revisionless def URL (urlNoRev) and
// its repo (as a hint for the string replace algorithm), and it adds
// the given revision (rev) to the URL. It is a special fastpath version
// because it is called very frequently during rendering of BlobLine.
function fastInsertRevIntoDefURL(urlNoRev: string, repo: string, rev: string): string {
	if (!rev) return urlNoRev;

	const prefix = `/${repo}/-/`;
	const repl = `/${repo}@${rev}/-/`;
	if (urlNoRev.startsWith(prefix)) return `${repl}${urlNoRev.slice(prefix.length)}`;
	return urlNoRev;
}

class BlobLine extends Component {
	componentDidMount(nextProps, nextState) {
		if (this.state.onMount) this.state.onMount();
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.path = props.path || null;
		state.textSize = props.textSize || "normal";

		// Update ownAnnURLs when they change.
		if (state.annotations !== props.annotations) {
			state.annotations = props.annotations;
			if (state.annotations && state.annotations.length) {
				state.ownAnnURLs = {};
				state.annotations.forEach((ann) => {
					if (ann.URL) state.ownAnnURLs[ann.URL] = true;
					if (ann.URLs) ann.URLs.forEach((url) => state.ownAnnURLs[url] = true);
				});
			} else {
				state.ownAnnURLs = null;
			}
		}

		// Filter to improve perf.
		state.highlightedDef = state.ownAnnURLs && state.ownAnnURLs[props.highlightedDef] ? props.highlightedDef : null;
		state.highlightedDefObj = state.highlightedDef ? props.highlightedDefObj : null;
		const activeDefURL = fastURLToRepoDef(props.activeDefRepo || state.repo, null, props.activeDef);
		state.activeDefURL = activeDefURL && state.ownAnnURLs && state.ownAnnURLs[activeDefURL] ? activeDefURL : null;

		state.lineNumber = props.lineNumber || null;
		state.showLineNumber = props.showLineNumber || false;
		state.startByte = props.startByte;
		state.contents = props.contents;
		state.selected = Boolean(props.selected);
		state.className = props.className || "";
		state.onMount = props.onMount || null;
	}

	_hasLink(content) {
		if (!(content instanceof Array)) return false;
		return content.some(item => {
			if (item.type === "a") {
				return true;
			}
			let props = item.props || {};
			return this._hasLink(props.children);
		});
	}

	_annotate() {
		let i = 0;
		return fromUtf8(annotate(this.state.contents, this.state.startByte, this.state.annotations, (ann, content) => {
			i++;

			const annURLs = (ann.URL ? [ann.URL] : ann.URLs) || null;

			// annRevURLs are the ann's URLs with the correct revision added. The raw
			// ann.URL/ann.URLs values, if they are def URLs, never contain revs.
			let annRevURLs = annURLs ? annURLs.map((url) => fastInsertRevIntoDefURL(url, this.state.repo, this.state.rev)) : null;

			// If ann.URL is an absolute URL with scheme http or https, create an anchor with a link to the URL (e.g., an
			// external URL to Mozilla's CSS reference documentation site.
			if (annURLs && isExternalLink(annURLs[0])) {
				let isHighlighted = this.state.highlightedDef === annURLs[0];
				return (
					<a
						className={classNames(ann.Class, {
							[s.highlightedAnn]: isHighlighted && (!this.state.highlightedDefObj || !this.state.highlightedDefObj.Error),
						})}
						target="_blank"
						href={annURLs[0]}
						onMouseOver={() => Dispatcher.Stores.dispatch(new DefActions.Hovering({repo: this.state.repo, commit: this.state.commitID, file: this.state.path, line: this.state.lineNumber - 1, character: ann.StartByte - this.state.startByte}))}
						onMouseOut={() => Dispatcher.Stores.dispatch(new DefActions.Hovering(null))}
						key={i}>
						{simpleContentsString(content)}
					</a>
				);
			}

			// ensure there are no links inside content to make ReactJS happy
			// otherwise incorrect DOM is built (a > .. > a)
			if (annURLs && annRevURLs && !this._hasLink(content)) {
				let isHighlighted = annURLs.includes(this.state.highlightedDef);
				return (
					<Link
						className={classNames(ann.Class, {
							[s.highlightedAnn]: isHighlighted && (!this.state.highlightedDefObj || !this.state.highlightedDefObj.Error),

							// disabledAnn is an ann that you can't click on (possibly a broken ref).
							[s.disabledAnn]: isHighlighted && (this.state.highlightedDefObj && this.state.highlightedDefObj.Error),

							[s.activeAnn]: annURLs.includes(this.state.activeDefURL),
						})}
						to={annRevURLs[0]}
						onMouseOver={() => Dispatcher.Stores.dispatch(new DefActions.Hovering({repo: this.state.repo, commit: this.state.commitID, file: this.state.path, line: this.state.lineNumber - 1, character: ann.StartByte - this.state.startByte}))}
						onMouseOut={() => Dispatcher.Stores.dispatch(new DefActions.Hovering(null))}
						onClick={(ev) => {
							if (ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey) return;
							// TODO: implement multiple defs menu if ann.URLs.length > 0 (more important for languages other than Go)
							if (this.state.highlightedDefObj && this.state.highlightedDefObj.Error) {
								// Prevent navigating to a broken ref or not-yet-loaded def.
								ev.preventDefault();
							}

							// Clear the def tooltip on click, or else it might be stuck
							// to the cursor if no corresponding Hovering(null) is dispatched.
							Dispatcher.Stores.dispatch(new DefActions.Hovering(null));
						}}
						key={i}>{fromUtf8(content)}</Link>
				);
			}
			return <span key={i} className={ann.Class}>{fromUtf8(content)}</span>;
		}));
	}

	render() {
		let contents = this.state.annotations ? this._annotate() : simpleContentsString(this.state.contents);

		// A single newline makes this line show up (correctly) as an empty line
		// when copied and pasted, instead of being omitted entirely.
		if (!contents) contents = "\n";

		return (
			<tr className={`${s.line} ${s[this.state.textSize]} ${this.state.className || ""}`}
				data-line={this.state.lineNumber}>
				{this.state.showLineNumber &&
					<td className={s.lineNumberCell} onClick={(event) => {
						if (event.shiftKey) {
							event.preventDefault();
							Dispatcher.Stores.dispatch(new BlobActions.SelectLineRange(this.state.repo, this.state.rev, this.state.path, this.state.lineNumber));
							return;
						}
					}}>
						<Link className={this.state.selected ? s.selectedLineNumber : s.lineNumber}
							to={`${urlToBlob(this.state.repo, this.state.rev, this.state.path)}#L${this.state.lineNumber}`} data-line={this.state.lineNumber} />
					</td>}

				<td className={`code ${this.state.selected ? s.selectedLineContent : s.lineContent}`}>
					{contents}
				</td>
			</tr>
		);
	}
}

BlobLine.propTypes = {
	lineNumber: (props, propName, componentName) => {
		let v = React.PropTypes.number(props, propName, componentName);
		if (v) return v;
		return null;
	},
	showLineNumber: React.PropTypes.bool,

	// Optional: for linking line numbers to the file they came from (e.g., in
	// ref snippets).
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	commitID: React.PropTypes.string,
	path: React.PropTypes.string,

	activeDef: React.PropTypes.string, // the def that the page is about

	// startByte is the byte position of the first byte of contents. It is
	// required if annotations are specified, so that the annotations can
	// be aligned to the contents.
	startByte: (props, propName, componentName) => {
		if (props.annotations) return React.PropTypes.number.isRequired(props, propName, componentName);
		return null;
	},
	contents: React.PropTypes.string,
	annotations: React.PropTypes.array,
	selected: React.PropTypes.bool,
	highlightedDef: React.PropTypes.string,
	className: React.PropTypes.string,
	onMount: React.PropTypes.func,
};

export default BlobLine;
