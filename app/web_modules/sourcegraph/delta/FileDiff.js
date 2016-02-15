import React from "react";

import DiffStatScale from "sourcegraph/delta/DiffStatScale";
import Hunk from "sourcegraph/delta/Hunk";
import router from "../../../script/routing/router";

class FileDiff extends React.Component {
	render() {
		let diff = this.props.diff;
		return (
			<div className="file-diff" id={this.props.id || ""}>
				<header>
					<DiffStatScale Stat={diff.Stats} />

					<span>{diff.OrigName === "/dev/null" ? diff.NewName : diff.OrigName}</span>
					{diff.NewName !== diff.OrigName && diff.OrigName !== "/dev/null" && diff.NewName !== "/dev/null" ? (
						<span> <i className="fa fa-long-arrow-right" /> {diff.NewName}</span>
					) : null}

					<div className="btn-group pull-right">
						{diff.OrigName !== "/dev/null" && <a className="button btn btn-default btn-xs" href={router.fileURL(this.props.baseRepo, this.props.baseRev, diff.OrigName)}>Original</a>}
						{diff.NewName !== "/dev/null" && <a className="button btn btn-default btn-xs" href={router.fileURL(this.props.headRepo, this.props.headRev, diff.NewName)}>New</a>}
					</div>
				</header>

				{diff.Hunks.map((hunk, i) => <Hunk key={i} hunk={hunk} />)}
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

	// id is the optional DOM ID, used for creating a URL ("...#F1")
	// that points to this specific file in a multi-file diff.
	id: React.PropTypes.string,
};
export default FileDiff;
