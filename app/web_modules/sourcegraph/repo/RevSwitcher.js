import React from "react";
import classNames from "classnames";
import Fuze from "fuse.js";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import * as router from "sourcegraph/util/router";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Component from "sourcegraph/Component";

function newFuzzyFinder(list) {
	return new Fuze(list.map((i) => i.Name), {
		distance: 1000,
		location: 0,
		threshold: 0.1,
	});
}

class RevSwitcher extends Component {
	constructor(props) {
		super(props);
		this.state = {
			open: false,
		};
		this._onToggleDropdown = this._onToggleDropdown.bind(this);
		this._onChangeQuery = this._onChangeQuery.bind(this);
		this._debouncedSetQuery = debounce((query) => {
			this.setState({query: query});
		}, 300, {leading: false, trailing: true});
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.branchesErr = this.state.branches ? this.state.branches.error(state.repo) : null;
		state.tagsErr = this.state.tags ? this.state.tags.error(state.repo) : null;
	}

	onStateTransition(prevState, nextState) {
		const becameOpen = nextState.open && nextState.open !== prevState.open;
		if (becameOpen || nextState.repo !== prevState.repo) {
			// Don't load the file list when the page loads until we become open.
			const initialLoad = !prevState.repo;
			if (!initialLoad || nextState.prefetch) {
				Dispatcher.asyncDispatch(new RepoActions.WantBranches(nextState.repo));
				Dispatcher.asyncDispatch(new RepoActions.WantTags(nextState.repo));
			}
		}

		["branches", "tags"].forEach((what) => {
			let nextItems = nextState[what] && nextState[what].list(nextState.repo);
			let prevItems = prevState[what] && prevState[what].list(prevState.repo);
			if (nextItems !== prevItems) {
				nextState[`${what}FuzzyFinder`] = nextItems ? newFuzzyFinder(nextItems) : null;
			}
			if (nextState[`${what}FuzzyFinder`] !== prevState[`${what}FuzzyFinder`] || nextState.query !== prevState.query) {
				nextState[`${what}Matches`] = (nextState.query && nextState[`${what}FuzzyFinder`]) ? nextState[`${what}FuzzyFinder`].search(nextState.query).map(i => nextItems[i]) : nextItems;
			}
		});
	}

	_loadingItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#">Loading {what}&hellip;</a></li>;
	}

	_errorItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#" className="text-danger">Error</a></li>;
	}

	_emptyItem(what) {
		return <li role="presentation" className="disabled"><a role="menuitem" tabIndex="-1" href="#">Nothing to show</a></li>;
	}

	_item(name, commitID) {
		let isCurrent = name === this.state.rev;
		return (
			<li key={`r${name}.${commitID}`} role="presentation" className={isCurrent && "current-rev"}>
				<a href={this._revSwitcherURL(name)} title={commitID}>
					{isCurrent && <i className="fa fa-caret-right"></i>}
					{name}
				</a>
			</li>
		);
	}

	// If path is not present, it means this is the rev switcher on commits page.
	_revSwitcherURL(rev) {
		switch (this.state.route) {
		case "tree":
			return router.tree(this.state.repo, rev, this.state.path);

		case "commits":
			return router.repoCommits(this.state.repo, rev);

		default:
			throw new Error(`can't generate URL for ${this.state.route}`);
		}
	}

	_onToggleDropdown() {
		this.setState({open: !this.state.open}, () => {
			if (this.state.open && this.refs.input) this.refs.input.focus();
		});
	}

	_onChangeQuery(ev) {
		if (this.refs.input) this._debouncedSetQuery(this.refs.input.value);
	}

	render() {
		let branches;
		if (this.state.branchesErr) {
			branches = this._errorItem("branches");
		} else if (this.state.branchesMatches === null) {
			branches = this._loadingItem("branches");
		} else if (this.state.branchesMatches.length === 0) {
			branches = this._emptyItem("branches");
		} else {
			branches = this.state.branchesMatches.map((b) => this._item(b.Name, b.Head));
		}

		let tags;
		if (this.state.tagsErr) {
			tags = this._errorItem("tags");
		} else if (this.state.tagsMatches === null) {
			tags = this._loadingItem("tags");
		} else if (this.state.tagsMatches.length === 0) {
			tags = this._emptyItem("tags");
		} else {
			tags = this.state.tagsMatches.map((tag) => this._item(tag.Name, tag.CommitID));
		}

		let cls = classNames("btn-group repo-rev-switcher", {
			open: this.state.open,
		});

		return (
			<div className={cls}>
				<button type="button" className="button btn btn-default dropdown-toggle"
					onClick={this._onToggleDropdown}>
					<i className="octicon octicon-git-branch"></i> {abbrevRev(this.props.rev)} <span className="caret"></span>
				</button>
				<div className={this.props.alignRight ? "dropdown-menu dropdown-menu-right" : "dropdown-menu"} role="menu">
					<div className="search-section">
						<input ref="input" type="text" className="form-control" placeholder="Find branch or tag" onChange={this._onChangeQuery}/>
					</div>
					<div role="presentation" className="divider"></div>
					<ul className="list-section">
						<li role="presentation" className="dropdown-header">Branches</li>
						{branches}
						<li role="presentation" className="divider"></li>
						<li role="presentation" className="dropdown-header">Tags</li>
						{tags}
					</ul>
				</div>
			</div>
		);
	}
}

// abbrevRev shortens rev if it is an absolute commit ID.
function abbrevRev(rev) {
	if (rev.length === 40) return rev.substring(0, 6);
	return rev;
}

RevSwitcher.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
	path: React.PropTypes.string.isRequired,
	route: React.PropTypes.oneOf(["tree", "commits"]).isRequired,

	// branches is RepoStore.branches.
	branches: React.PropTypes.object.isRequired,

	// tags is RepoStore.tags.
	tags: React.PropTypes.object.isRequired,

	// alignRight optionally allows the dropdown to be aligned right instead of the default alignment.
	alignRight: React.PropTypes.bool,
};

export default RevSwitcher;
