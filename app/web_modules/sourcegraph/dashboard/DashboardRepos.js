import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import moment from "moment";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchQuery: "",
			filter: null,
		};
		this._handleSearch = this._handleSearch.bind(this);
		this._selectFilter = this._selectFilter.bind(this);
		this._showRepo = this._showRepo.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleSearch(e) {
		this.setState(update(this.state, {
			searchQuery: {$set: e.target.value},
		}));
	}

	_selectFilter(filterValue) {
		this.setState(update(this.state, {
			filter: {$set: filterValue},
		}));
	}

	_showRepo(repo) {
		const isPrivate = Boolean(repo.Private);
		if (this.state.searchQuery && repo.URI.indexOf(this.state.searchQuery) === -1) {
			return false;
		}
		if (this.state.filter) {
			if (this.state.filter === "private" && !isPrivate) {
				return false;
			}
			if (this.state.filter === "public" && isPrivate) {
				return false;
			}
		}
		return true; // no filter; return all
	}

	render() {
		const toggles = [null, "private", "public"].map((filterValue, i) =>
			<button key={i}
				className={`btn btn-block toggle ${this.state.filter === filterValue ? "btn-primary" : "btn-default"}`}
				onClick={() => this._selectFilter(filterValue)}>
				<span className="toggle-label">{filterValue ? filterValue : "all"}</span>
			</button>
		);

		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<div className="toggles">
						<div className="btn-group">{toggles}</div>
					</div>
					<div className="search-bar">
						<input className="form-control search-input"
							placeholder="Search repositories"
							value={this.state.searchQuery}
							onChange={this._handleSearch}
							type="text" />
					</div>
				</div>
				<div className="repos-list panel-body">
					<div className="list-group">
						{this.state.repos.filter(this._showRepo).map((repo, i) => (
							<div className="list-group-item hover-pointer" key={i}
								onClick={() => window.location.href = repo.HTMLURL}>
								<div className="repo-header">
									<div className="repo-icon">
									</div>
									<h4>{repo.URI}</h4>
								</div>
								<div className="repo-body">
									<p className="description">{repo.Description}</p>
									<p className="updated">{`Updated ${moment(repo.UpdatedAt).fromNow()}`}</p>
								</div>
							</div>
						))}
					</div>
				</div>
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default DashboardRepos;
