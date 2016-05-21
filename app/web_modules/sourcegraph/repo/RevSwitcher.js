import React from "react";
import Fuze from "fuse.js";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import Component from "sourcegraph/Component";
import {Link} from "react-router";
import styles from "./styles/RevSwitcher.css";
import {TriangleDownIcon, CheckIcon} from "sourcegraph/components/Icons";
import Input from "sourcegraph/components/Input";
import CSSModules from "react-css-modules";
import {urlWithRev} from "sourcegraph/repo/routes";

function newFuzzyFinder(list) {
	return new Fuze(list.map((i) => i.Name), {
		distance: 1000,
		location: 0,
		threshold: 0.1,
	});
}

class RevSwitcher extends Component {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string.isRequired,
		repoObj: React.PropTypes.object,
		isCloning: React.PropTypes.bool.isRequired,

		// branches is RepoStore.branches.
		branches: React.PropTypes.object.isRequired,

		// tags is RepoStore.tags.
		tags: React.PropTypes.object.isRequired,

		// srclibDataVersions is TreeStore.srclibDataVersions.
		srclibDataVersions: React.PropTypes.object.isRequired,

		// to construct URLs
		routes: React.PropTypes.array.isRequired,
		routeParams: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			open: false,
		};
		this._closeDropdown = this._closeDropdown.bind(this);
		this._onToggleDropdown = this._onToggleDropdown.bind(this);
		this._onChangeQuery = this._onChangeQuery.bind(this);
		this._onClickOutside = this._onClickOutside.bind(this);
		this._onKeydown = this._onKeydown.bind(this);
		this._debouncedSetQuery = debounce((query) => {
			this.setState({query: query});
		}, 150, {leading: true, trailing: true});
	}

	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
		if (typeof document !== "undefined") {
			document.addEventListener("click", this._onClickOutside);
			document.addEventListener("keydown", this._onKeydown);
		}
	}

	componentWillUnmount() {
		if (super.componentWillUnmount) super.componentWillUnmount();
		if (typeof document !== "undefined") {
			document.removeEventListener("click", this._onClickOutside);
			document.removeEventListener("keydown", this._onKeydown);
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.branchesErr = state.branches ? state.branches.error(state.repo) : null;
		state.tagsErr = state.tags ? state.tags.error(state.repo) : null;

		state.srclibDataVersion = state.srclibDataVersions ? state.srclibDataVersions.get(state.repo, state.commitID) : null;

		// effectiveRev is the rev from the URL, or else the repo's default branch.
		state.effectiveRev = state.rev || (state.repoObj && !state.repoObj.Error ? state.repoObj.DefaultBranch : null);
	}

	onStateTransition(prevState, nextState) {
		const becameOpen = nextState.open && nextState.open !== prevState.open;
		if (becameOpen || nextState.repo !== prevState.repo) {
			// Don't load when page loads until we become open.
			const initialLoad = !prevState.repo && !nextState.open;
			if (!initialLoad || nextState.prefetch) {
				Dispatcher.Backends.dispatch(new RepoActions.WantBranches(nextState.repo));
				Dispatcher.Backends.dispatch(new RepoActions.WantTags(nextState.repo));
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.commitID, null));
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
		return <li role="presentation" styleName="disabled">Loading {what}&hellip;</li>;
	}

	_errorItem(what) {
		return <li role="presentation" styleName="disabled">Error</li>;
	}

	_emptyItem(what) {
		return <li role="presentation" styleName="disabled">Nothing to show</li>;
	}

	_item(name, commitID) {
		let isCurrent = name === this.state.effectiveRev;

		const unindexed = this.state.srclibDataVersion && !this.state.srclibDataVersion.CommitID;
		const commitsBehind = this.state.srclibDataVersion && !this.state.srclibDataVersion.Error ? this.state.srclibDataVersion.CommitsBehind : 0;

		return (
			<li key={`r${name}.${commitID}`}
				styleName={isCurrent ? "item-current" : "item"}>
				<Link to={this._revSwitcherURL(name)} title={commitID}
					styleName={isCurrent ? "item-content-current" : "item-content"}
					onClick={this._closeDropdown}>
					<CheckIcon styleName={isCurrent ? "icon" : "icon-hidden"} /> <span styleName="item-name">{abbrevRev(name)}</span>
					{isCurrent && commitsBehind ? <span styleName="detail">{commitsBehind} commit{commitsBehind !== 1 && "s"} ahead of index</span> : null}
					{isCurrent && unindexed ? <span styleName="detail">not indexed</span> : null}
				</Link>
			</li>
		);
	}

	_closeDropdown(ev) {
		// HACK: If the user clicks to a rev that they have already loaded all
		// of the data for, the transition occurs synchronously and the dropdown
		// does not close for some reason. Bypassing this.setState and setting it
		// directly fixes this issue.
		this.state.open = false; // eslint-disable-line react/no-direct-mutation-state
		this.setState({open: false});
	}

	// If path is not present, it means this is the rev switcher on commits page.
	_revSwitcherURL(rev) {
		return urlWithRev(this.state.routes, this.state.routeParams, rev);
	}

	_onToggleDropdown(ev) {
		ev.preventDefault();
		ev.stopPropagation();
		this.setState({open: !this.state.open}, () => {
			if (this.state.open && this._input) this._input.focus();
		});
	}

	_onChangeQuery(ev) {
		if (this._input) this._debouncedSetQuery(this._input.value);
	}

	// _onClickOutside causes clicks outside the menu to close the menu.
	_onClickOutside(ev) {
		if (!this.state.open) return;
		if (this._wrapper && !this._wrapper.contains(ev.target)) this.setState({open: false});
	}

	// _onKeydown causes ESC to close the menu.
	_onKeydown(ev) {
		if (!this.state.open) return;
		if (ev.keyCode === 27 /* ESC */) {
			this.setState({open: false});
		}
	}

	render() {
		// Hide if cloning the repo, since we require the user to hard-reload. Seeing
		// the RevSwitcher would confuse them.
		if (this.state.isCloning) return null;

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

		let currentItem;
		(this.state.branchesMatches || []).forEach((b) => {
			if (b.Name === this.state.effectiveRev) currentItem = b;
		});
		(this.state.tagsMatches || []).forEach((t) => {
			if (t.Name === this.state.effectiveRev) currentItem = t;
		});

		let title;
		if (this.state.rev) title = `Viewing revision: ${abbrevRev(this.state.rev)}`;
		else if (this.state.srclibDataVersion && this.state.srclibDataVersion.CommitID) title = `Viewing last-built revision on default branch: ${this.state.commitID ? abbrevRev(this.state.commitID) : ""}`;
		else title = `Viewing revision: ${abbrevRev(this.state.commitID)} (not indexed)`;

		return (
			<div styleName="wrapper"
				ref={(e) => this._wrapper = e}>
				<span styleName="toggle"
					title={title}
					onClick={this._onToggleDropdown}>
					<TriangleDownIcon />
				</span>
				<div role="menu"
					styleName={this.state.open ? "dropdown-menu-open" : "dropdown-menu-closed"}>
					<div styleName="search-section">
						<Input block={true}
							domRef={(e) => this._input = e}
							type="text"
							styleName="input"
							placeholder="Find branch or tag"
							onChange={this._onChangeQuery}/>
					</div>
					<div role="presentation" styleName="divider"></div>
					<ul styleName="list-section">
						{/* Show the current one at the top if it wouldn't otherwise be shown. */}
						{!currentItem && !this.state.query && this._item(this.state.rev, this.state.commitID)}
						<li role="presentation" styleName="dropdown-header">Branches</li>
						{branches}
						<li role="presentation" styleName="divider"></li>
						<li role="presentation" styleName="dropdown-header">Tags</li>
						{tags}
					</ul>
				</div>
			</div>
		);
	}
}

// abbrevRev shortens rev if it is an absolute commit ID.
function abbrevRev(rev) {
	if (rev.length === 40) return rev.substring(0, 12);
	return rev;
}

export default CSSModules(RevSwitcher, styles);
