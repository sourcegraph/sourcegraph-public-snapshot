import React from "react";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import {atob} from "abab";
import Component from "sourcegraph/Component";
import DiffStatScale from "sourcegraph/delta/DiffStatScale";
import Dispatcher from "sourcegraph/Dispatcher";
import Hunk from "sourcegraph/delta/Hunk";
import router from "../../../script/routing/router";
import fileLines from "sourcegraph/util/fileLines";
import {isDevNull} from "sourcegraph/delta/util";

class FileDiff extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	onStateTransition(prevState, nextState) {
		if (nextState.baseRepo !== prevState.baseRepo || nextState.baseRev !== prevState.baseRev || nextState.diff.OrigName !== prevState.diff.OrigName) {
			if (!isDevNull(nextState.diff.OrigName)) Dispatcher.asyncDispatch(new BlobActions.WantAnnotations(nextState.baseRepo, nextState.baseRev, "", nextState.diff.OrigName));
		}
		if (nextState.headRepo !== prevState.headRepo || nextState.headRev !== prevState.headRev || nextState.diff.NewName !== prevState.diff.NewName) {
			if (!isDevNull(nextState.diff.NewName)) Dispatcher.asyncDispatch(new BlobActions.WantAnnotations(nextState.headRepo, nextState.headRev, "", nextState.diff.NewName));
		}
	}

	_groupAnnotationsByHunk(hunks) {
		let baseAnns = this.state.annotations.get(this.state.baseRepo, this.state.baseRev, "", this.state.diff.OrigName);
		let headAnns = this.state.annotations.get(this.state.headRepo, this.state.headRev, "", this.state.diff.NewName);
		return hunks.map((hunk) => {
			let baseStart = baseAnns ? baseAnns.LineStartBytes[hunk.OrigStartLine - 1] : null;
			let baseEnd = baseAnns ? baseAnns.LineStartBytes[hunk.OrigStartLine + hunk.OrigLines - 1] : null;
			let hunkBaseAnns = baseAnns ? baseAnns.Annotations.filter((ann) => ann.StartByte >= baseStart && ann.EndByte < baseEnd) : [];

			let headStart = headAnns ? headAnns.LineStartBytes[hunk.NewStartLine - 1] : null;
			let headEnd = headAnns ? headAnns.LineStartBytes[hunk.NewStartLine + hunk.NewLines - 1] : null;
			let hunkHeadAnns = headAnns ? headAnns.Annotations.filter((ann) => ann.StartByte >= headStart && ann.EndByte < headEnd) : [];

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
					if (headAnns && (!baseAnns || headAnns.Annotations.length > baseAnns.Annotations.length)) addLineAnns(hunkHeadAnns, headAnns.LineStartBytes[newLine - 1], line.length);
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
		let hunkAnns = this._groupAnnotationsByHunk(diff.Hunks);
		return (
			<div className="file-diff" id={this.props.id || ""}>
				<header>
					<DiffStatScale Stat={diff.Stats} />

					<span>{isDevNull(diff.OrigName) ? diff.NewName : diff.OrigName}</span>
					{diff.NewName !== diff.OrigName && !isDevNull(diff.OrigName) && !isDevNull(diff.NewName) ? (
						<span> <i className="fa fa-long-arrow-right" /> {diff.NewName}</span>
					) : null}

					<div className="btn-group pull-right">
						{!isDevNull(diff.OrigName) && <a className="button btn btn-default btn-xs" href={router.fileURL(this.props.baseRepo, this.props.baseRev, diff.OrigName)}>Original</a>}
						{!isDevNull(diff.NewName) && <a className="button btn btn-default btn-xs" href={router.fileURL(this.props.headRepo, this.props.headRev, diff.NewName)}>New</a>}
					</div>
				</header>

				{diff.Hunks.map((hunk, i) => (
					<Hunk
						key={i}
						hunk={hunk}
						baseRepo={this.props.baseRepo}
						baseRev={this.props.baseRev}
						basePath={diff.OrigName}
						headRepo={this.props.headRepo}
						headRev={this.props.headRev}
						headPath={diff.NewName}
						annotations={hunkAnns[i]} />
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
};
export default FileDiff;
