var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var router = require("../routing/router");
var Fuse = require("fuse.js");

var RepoRevSwitcher = React.createClass({

	propTypes: {
		// route specifies the route to use when switching branches. Its value
		// represents the function that will be called in the 'router' module
		// with the "URL" string appended at the end.
		// Example:
		// a value of "file" will cause the route to become router.fileURL(repo, rev, path).
		route: React.PropTypes.string,

		// alignRight optionally allows the dropdown to be aligned right instead of the default alignment.
		alignRight: React.PropTypes.bool,
	},

	getInitialState() { return {startedLoading: false}; },

	loadItems() {
		if (this.state.startedLoading) return;
		this.setState({startedLoading: true});
		this.loadItemsOfType("branches");
		this.loadItemsOfType("tags");
	},

	loadItemsOfType(what) {
		$.get(
			`/api/repos/${this.props.repoSpec}/.${what}`
		).success(function(resp) {
			resp = resp || [];
			var newState = {};
			newState[what] = resp;
			newState[`${what}Error`] = false;
			this.setState(newState);
		}.bind(this)).error(function(err) {
			console.error(err);
			var newState = {};
			newState[`${what}Error`] = true;
			this.setState(newState);
		}.bind(this));
	},

	makeLoadingItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#">Loading {what}&hellip;</a></li>;
	},

	makeErrorItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#" className="text-danger">Error</a></li>;
	},

	makeEmptyItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#">none</a></li>;
	},

	makeItem(name, commitID) {
		var isCurrent = name === this.props.rev;
		return (
			<li key={`r${name}.${commitID}`} role="presentation" className={isCurrent && "current-rev"}>
				<a onClick={this._onClickBranch.bind(this, this.props.repoSpec, name, this.props.path)} href={this._revSwitcherURL(name)} title={commitID}>
					{isCurrent && <i className="fa fa-caret-right"></i>}
					{name}
				</a>
			</li>
		);
		// TODO(sqs): add back the build indicator here. the css was tough to get right in the dropdown.
		// <RepoBuildIndicator RepoURI={this.props.repoSpec} Rev={commitID}/>
	},

	// Filter the branches or tags that will be shown in the menu via fuzzy search.
	filterItems(items) {
		if (!this.state.searchPattern) return items;
		var f = new Fuse(items, {
			keys: ["Name"],
			threshold: 0.4,
			distance: 100,
			location: 0,
		});
		return f.search(this.state.searchPattern);
	},

	// If path is not present, it means this is the rev switcher on commits page.
	_revSwitcherURL(rev) {
		var repo = this.props.repoSpec,
			path = this.props.path,
			fn = router[`${this.props.route}URL`];

		if (typeof fn === "function") {
			return fn(repo, rev, path);
		}

		return router.fileURL(repo, rev, path);
	},

	_onToggleDropdown() {
		var $el = $(ReactDOM.findDOMNode(this));
		if ($el.hasClass("open")) {
			var searchInput = ReactDOM.findDOMNode(this.refs.itemSearch);
			$(searchInput).focus();
		}

		this.loadItems();
	},

	_onClickBranch(repoSpec, name, path, e) {
		if (typeof this.props.onBranchSelect === "function") {
			this.props.onBranchSelect(repoSpec, name, path, e);
			e.preventDefault();
		}
	},

	_onUpdateSearchPattern() {
		this.setState({searchPattern: ReactDOM.findDOMNode(this.refs.itemSearch).value});
	},


	render() {
		var branches;
		if (this.state.branchesError) {
			branches = this.makeErrorItem("branches");
		} else if (typeof this.state.branches === "undefined") {
			branches = this.makeLoadingItem("branches");
		} else if (!this.state.branches.Branches || this.state.branches.Branches.length === 0) {
			branches = this.makeEmptyItem("branches");
		} else {
			var filteredBranches = this.filterItems(this.state.branches.Branches);
			if (filteredBranches.length === 0) {
				branches = this.makeEmptyItem("branches");
			} else {
				branches = filteredBranches.map(function(b) {
					return this.makeItem(b.Name, b.Head);
				}.bind(this));
			}
		}

		var tags;
		if (this.state.tagsError) {
			tags = this.makeErrorItem("tags");
		} else if (typeof this.state.tags === "undefined") {
			tags = this.makeLoadingItem("tags");
		} else if (!this.state.tags.Tags || this.state.tags.Tags.length === 0) {
			tags = this.makeEmptyItem("tags");
		} else {
			var filteredTags = this.filterItems(this.state.tags.Tags);
			if (filteredTags.length === 0) {
				tags = this.makeEmptyItem("tags");
			} else {
				tags = filteredTags.map(function(tag) {
					return this.makeItem(tag.Name, tag.CommitID);
				}.bind(this));
			}
		}

		return (
			<div className="btn-group repo-rev-switcher">
				<button type="button" className="button btn btn-default btn-sm dropdown-toggle" data-toggle="dropdown" onClick={this._onToggleDropdown}>
					<i className="octicon octicon-git-branch"></i> {this.props.label ? <span className="label">{this.props.label}</span> : null} {this.props.rev} <span className="caret"></span>
				</button>
				<ul className={this.props.alignRight ? "dropdown-menu dropdown-menu-right" : "dropdown-menu"} role="menu">
					<li className="search-section">
						<input ref="itemSearch" type="text" className="form-control" placeholder="Find branch or tag" onChange={this._onUpdateSearchPattern}/>
					</li>
					<li role="presentation" className="divider"></li>
					<ul className="list-section">
						<li role="presentation" className="dropdown-header">Branches</li>
						{branches}
						<li role="presentation" className="divider"></li>
						<li role="presentation" className="dropdown-header">Tags</li>
						{tags}
					</ul>
				</ul>
			</div>
		);
	},
});

module.exports = RepoRevSwitcher;
