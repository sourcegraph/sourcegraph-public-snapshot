import React from "react";
import styles from "sourcegraph/delta/styles/DiffStatScale.css";
import CSSModules from "react-css-modules";

class DiffStatScale extends React.Component {
	render() {
		if (!this.props.Stat) return null;

		let ds = this.props.Stat;
		let x;
		let sum = ds.Added + ds.Changed + ds.Deleted;

		let sds = {
			Added: ds.Added,
			Changed: ds.Changed,
			Deleted: ds.Deleted,
			ScaledAdded: ds.Added,
			ScaledChanged: ds.Changed,
			ScaledDeleted: ds.Deleted,
		};

		let size = this.props.Size || 5;
		if (sum > size) {
			x = size / sum;
			sds.ScaledAdded = parseInt(sds.ScaledAdded * x, 10);
			if (sds.ScaledAdded === 0 && sds.Added !== 0) {
				sds.ScaledAdded = 1;
			}
			sds.ScaledDeleted = parseInt(sds.ScaledDeleted * x, 10);
			if (sds.ScaleddDeleted === 0 && sds.Deleted !== 0) {
				sds.ScaledDeleted = 1;
			}
			sds.ScaledChanged = parseInt(sds.ScaledChanged * x, 10);
			if (sds.ScaledChanged === 0 && sds.Changed !== 0) {
				sds.ScaledChanged = 1;
			}
		}

		let filler = 0;
		if (sds.ScaledAdded + sds.ScaledChanged + sds.ScaledDeleted < size) {
			filler = size - (sds.ScaledAdded + sds.ScaledChanged + sds.ScaledDeleted);
		}

		function bar(width) {
			let s = "";
			for (let i = 0; i < width; i++) {
				s += "\u25A0";
			}
			return s;
		}

		return (
			<span styleName="diff-stat-scale">
				<span styleName="stat-added">{bar(sds.ScaledAdded)}</span>
				<span styleName="stat-changed">{bar(sds.ScaledChanged)}</span>
				<span styleName="stat-deleted">{bar(sds.ScaledDeleted)}</span>
				{filler > 0 ? (
					<span styleName="stat-filler">{bar(filler)}</span>
				) : null}
			</span>
		);
	}
}
DiffStatScale.propTypes = {
	Stat: React.PropTypes.shape({
		Added: React.PropTypes.number,
		Changed: React.PropTypes.number,
		Deleted: React.PropTypes.number,
	}).isRequired,
	Size: React.PropTypes.number,
};
export default CSSModules(DiffStatScale, styles);
