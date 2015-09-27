var React = require("react");
var router = require("../routing/router");
var $ = require("jquery");

var TreeEntryDefs = React.createClass({
	getInitialState() {
		return {defs: []};
	},

	componentDidMount() {
		var self = this;
		this.getChildDefs(function(defs) {
			self.setState({defs: defs});
		});
	},

	getChildDefs(cont) {
		var self = this;

		var repo = self.props.repo;
		var commit = self.props.commit;
		var path = self.props.path;
		var isFile = self.props.isFile;

		var q = {
			Direction: "desc",
			Exported: true,
			PerPage: 6,
			RepoRevs: `${repo}@${commit}`,
			Sort: "def_len",
		};

		if (isFile) {
			q.File = path;
		} else {
			q.FilePathPrefix = path;
		}

		$.ajax({
			url: "/api/.defs",
			data: q,
			success(d) {
				cont(d.Defs || []);
			},
			error() {
				cont([]);
			},
		});
	},

	render() {
		var self = this;

		var repo = self.props.repo;
		var rev = self.props.rev;
		var path = self.props.path;

		var defs = self.state.defs || [];

		var more = (defs.length !==0) ? <li className="more"><code><a href={router.fileURL(repo, rev, path)}>&hellip;</a></code></li> : null;

		return (
			<ul className="defs list-inline">
				{defs.map(function(d) {
					var href = router.defURL(repo, rev, d.UnitType, d.Unit, d.Path);

					if (d.Kind === "package") {
						return null;
					}

					return (
						<li key={d.TreePath}>
							<code>
								<a className="defn-popover" href={href}>
									{d.Name}
								</a>
							</code>
						</li>
					);
				})}
				{more}
			</ul>);
	},
});

module.exports = TreeEntryDefs;
