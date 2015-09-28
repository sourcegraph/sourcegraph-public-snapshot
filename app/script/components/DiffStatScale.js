var React = require("react");

var DiffStatScale = React.createClass({
	getDefaultProps() {
		return {
			Stat: {Changed: 0, Added: 0, Deleted: 0},
			Size: 5,
		};
	},

	render() {
		var ds = this.props.Stat;
		var x;
		var sum = ds.Added + ds.Changed + ds.Deleted;

		var sds = {
			Added: ds.Added,
			Changed: ds.Changed,
			Deleted: ds.Deleted,
			ScaledAdded: ds.Added,
			ScaledChanged: ds.Changed,
			ScaledDeleted: ds.Deleted,
		};

		if (sum > this.props.Size) {
			x = this.props.Size / sum;
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

		var filler = 0;
		if (sds.ScaledAdded + sds.ScaledChanged + sds.ScaledDeleted < this.props.Size) {
			filler = this.props.Size - (sds.ScaledAdded + sds.ScaledChanged + sds.ScaledDeleted);
		}

		function bar(width) {
			var s = "";
			for (var i = 0; i < width; i++) {
				s += "\u25A0";
			}
			return s;
		}

		return (
			<span className="diff-stat-scale">
				<span className="stat-added">{bar(sds.ScaledAdded)}</span>
				<span className="stat-changed">{bar(sds.ScaledChanged)}</span>
				<span className="stat-deleted">{bar(sds.ScaledDeleted)}</span>
				{filler > 0 ? (
					<span className="stat-filler">{bar(filler)}</span>
				) : null}
			</span>
		);
	},
});

module.exports = DiffStatScale;
