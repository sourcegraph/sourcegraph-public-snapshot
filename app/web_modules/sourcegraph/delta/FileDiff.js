import React from "react";
import styles from "sourcegraph/delta/styles/FileDiff.css";
import CSSModules from "react-css-modules";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {atob} from "abab";
import Component from "sourcegraph/Component";
import DiffStatScale from "sourcegraph/delta/DiffStatScale";
import Dispatcher from "sourcegraph/Dispatcher";
import Hunk from "sourcegraph/delta/Hunk";
import {urlToBlob} from "sourcegraph/blob/routes";
import fileLines from "sourcegraph/util/fileLines";
import {isDevNull} from "sourcegraph/delta/util";

class FileDiff extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);

		let baseAnns = state.annotations.get(state.baseRepo, state.baseRev, state.diff.OrigName);
		let headAnns = state.annotations.get(state.headRepo, state.headRev, state.diff.NewName);
		if (!state.hunkAnns || baseAnns !== state.baseAnns || headAnns !== state.headAnns) {
			if (state.diff.Hunks && state.hunkAnns && state.baseAnns && state.headAnns) {
				state.hunkAnns = this._groupAnnotationsByHunk(state.diff.Hunks, state.baseAnns, state.headAnns);
			} else {
				state.hunkAnns = null;
			}
		}
		state.baseAnns = baseAnns;
		state.headAnns = headAnns;

		// We've already pulled the info we need from annotations. Remove it on state so we don't
		// rerender each time it is updated when it's not relevant to us.
		delete state.annotations;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.baseRepo !== prevState.baseRepo || nextState.baseRev !== prevState.baseRev || nextState.diff !== prevState.diff) {
			if (nextState.diff && nextState.baseRepo && nextState.baseRev && !isDevNull(nextState.diff.OrigName)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(nextState.baseRepo, nextState.baseRev, nextState.diff.OrigName));
			}
		}
		if (nextState.headRepo !== prevState.headRepo || nextState.headRev !== prevState.headRev || nextState.diff !== prevState.diff) {
			if (nextState.headRepo && nextState.headRev && nextState.diff && !isDevNull(nextState.diff.NewName)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(nextState.headRepo, nextState.headRev, nextState.diff.NewName));
			}
		}
	}

	_groupAnnotationsByHunk(hunks, baseAnns, headAnns) {
		return hunks.map((hunk) => {
			let baseStart = baseAnns ? baseAnns.LineStartBytes[hunk.OrigStartLine - 1] : null;
			let baseEnd = baseAnns ? baseAnns.LineStartBytes[hunk.OrigStartLine + hunk.OrigLines - 1] : null;
			let hunkBaseAnns = baseAnns ? (baseAnns.Annotations || []).filter((ann) => ann.StartByte >= baseStart && ann.EndByte < baseEnd) : [];

			let headStart = headAnns ? headAnns.LineStartBytes[hunk.NewStartLine - 1] : null;
			let headEnd = headAnns ? headAnns.LineStartBytes[hunk.NewStartLine + hunk.NewLines - 1] : null;
			let hunkHeadAnns = headAnns ? (headAnns.Annotations || []).filter((ann) => ann.StartByte >= headStart && ann.EndByte < headEnd) : [];

			let resultAnns = [];
			let lines = fileLines(atob(hunk.Body));
			let origLine = hunk.OrigStartLine;
			let newLine = hunk.NewStartLine;
			let hunkLineStartByte = 0;

			let addLineAnns = (hunkHeadOrBaseAnns, lineStartByte, lineLengthIncludingPrefix) => {
				const lineEndByte = lineStartByte + lineLengthIncludingPrefix - 1; // remove 1-char prefix
				hunkHeadOrBaseAnns.forEach((ann) => {
					if (ann.StartByte >= lineStartByte && ann.EndByte <= lineEndByte) {
						resultAnns.push(Object.assign({}, ann, {
							// Add 1 to each to account for the 1-char prefix.
							StartByte: hunkLineStartByte + (ann.StartByte - lineStartByte) + 1,
							EndByte: hunkLineStartByte + (ann.EndByte - lineStartByte) + 1,
						}));
					}
				});
			};

			lines.forEach((line) => {
				switch (line[0]) {
				case "+":
					if (headAnns) addLineAnns(hunkHeadAnns, headAnns.LineStartBytes[newLine - 1], line.length);
					newLine++;
					break;

				case "-":
					if (baseAnns) addLineAnns(hunkBaseAnns, baseAnns.LineStartBytes[origLine - 1], line.length);
					origLine++;
					break;

				case " ":
					if (headAnns && (!baseAnns || !baseAnns.Annotations || (headAnns.Annotations && headAnns.Annotations.length > baseAnns.Annotations.length))) addLineAnns(hunkHeadAnns, headAnns.LineStartBytes[newLine - 1], line.length);
					else if (baseAnns) addLineAnns(hunkBaseAnns, baseAnns.LineStartBytes[origLine - 1], line.length);
					origLine++;
					newLine++;
					break;
				}
				hunkLineStartByte += line.length + 1; // include 1-char trailing newline
			});

			return resultAnns;
		});
	}

	render() {
		let diff = this.props.diff;
		return (
			<div styleName="container" id={this.props.id || ""}>
				<header styleName="header">
					<div styleName="info">
						<DiffStatScale Stat={diff.Stats} />
						<span styleName="name">
							<span>{isDevNull(diff.OrigName) ? diff.NewName : diff.OrigName}</span>
							{diff.NewName !== diff.OrigName && !isDevNull(diff.OrigName) && !isDevNull(diff.NewName) ? (
								<span> &rarr; {diff.NewName}</span>
							) : null}
						</span>
					</div>
					<div styleName="actions">
						{!isDevNull(diff.OrigName) && <a styleName="action" href={urlToBlob(this.props.baseRepo, this.props.baseRev, diff.OrigName)}>Original</a>}
						{!isDevNull(diff.NewName) && <a styleName="action" href={urlToBlob(this.props.headRepo, this.props.headRev, diff.NewName)}>New</a>}
					</div>
				</header>

				{diff.Hunks && diff.Hunks.map((hunk, i) => (
					<Hunk
						key={i}
						hunk={hunk}
						baseRepo={this.props.baseRepo}
						baseRev={this.props.baseRev}
						basePath={diff.OrigName}
						headRepo={this.props.headRepo}
						headRev={this.props.headRev}
						headPath={diff.NewName}
						highlightedDef={this.state.highlightedDef}
						highlightedDefObj={this.state.highlightedDefObj}
						annotations={this.state.hunkAnns ? this.state.hunkAnns[i] : null} />
				))}
			</div>
		);
	}
}
FileDiff.propTypes = {
	diff: React.PropTypes.object.isRequired,
	baseRepo: React.PropTypes.string.isRequired,
	baseRev: React.PropTypes.string.isRequired,
	headRepo: React.PropTypes.string.isRequired,
	headRev: React.PropTypes.string.isRequired,

	// annotations is BlobStore.annotations.
	annotations: React.PropTypes.object,

	// id is the optional DOM ID, used for creating a URL ("...#F1")
	// that points to this specific file in a multi-file diff.
	id: React.PropTypes.string,

	highlightedDef: React.PropTypes.string,
	highlightedDefObj: React.PropTypes.object,
};
export default CSSModules(FileDiff, styles);
