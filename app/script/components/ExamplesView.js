var React = require("react");
var router = require("../routing/router");
var CodeView = require("./CodeView");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var ExamplesModel = require("../stores/models/ExamplesModel");
var classNames = require("classnames");
var repoLink = require("sourcegraph/util/repoLink").default;

var ExampleView = React.createClass({

	propTypes: {
		onTokenFocus: React.PropTypes.func,
		onTokenBlur: React.PropTypes.func,
		onTokenClick: React.PropTypes.func,
		onChangePage: React.PropTypes.func,
		onShowSnippet: React.PropTypes.func,
		def: React.PropTypes.string,
		model: React.PropTypes.instanceOf(ExamplesModel).isRequired,
	},

	mixins: [ModelPropWatcherMixin],

	/**
	 * @description Returns a function that creates an action which directs the
	 * user to the passed page.
	 * @param {number} page - The page that the action should request.
	 * @returns {void}
	 * @private
	 */
	_changePage(page) {
		if (page === 0 || (this.state.lastPage && page > this.state.page) || this.state.loading) return null;

		return () => {
			if (typeof this.props.onChangePage === "function") {
				this.props.onChangePage(this.props.def, page);
				this.setState({page: page});
			}
		};
	},

	render() {
		// Server-side error when processing request.
		if (this.state.error) {
			return (
				<div className="error">
					<i className="fa fa-exclamation-triangle"></i>
					A server error occurred while fetching examples.
				</div>
			);
		}
		// Nothing loaded yet.
		if (typeof this.state.example === "undefined") {
			return null;
		}
		// No examples found.
		if (this.state.example === null) {
			return <i className="noExamples">No usage examples found.</i>;
		}

		var ex = this.state.example,
			defUrl = router.defURL(ex.DefRepo, ex.CommitID, ex.DefUnitType, ex.DefUnit, ex.DefPath),
			s = SnippetToBreadcrumb(ex.Repo, ex.CommitID, ex.File, ex.StartLine, ex.EndLine, defUrl, this.props.onShowSnippet);

		var leftClasses = classNames({
			"fa": true,
			"fa-chevron-circle-left": true,
			"btnNav": true,
			"disabled": this.state.page === 1,
		});

		var rightClasses = classNames({
			"fa": true,
			"fa-chevron-circle-right": true,
			"btnNav": true,
			"disabled": this.state.lastPage,
		});

		var loading = this.state.loading || this.props.loading;

		return (
			<div className="example">
				<header>
					<div className="pull-right">{repoLink(ex.Repo)}</div>
					<nav>
						<a onClick={this._changePage(this.state.page-1)} className={leftClasses}></a>
						<a onClick={this._changePage(this.state.page+1)} className={rightClasses}></a>
					</nav>
					{s}
					{loading ? <i className="fa fa-spinner fa-spin"></i> : null}
				</header>

				<div className="body">
					<CodeView
						{...this.props}
						lineNumbers={false}
						loading={loading}
						model={this.state.codeModel}
						theme="default" />
				</div>

				<footer>
					<a target="_blank" href={`${defUrl}/.examples`} className="pull-right">
						<i className="fa fa-eye" /> View all
					</a>
				</footer>
			</div>
		);
	},
});

module.exports = ExampleView;

// TODO(gbbr): This should be a React component.
// SnippetToBreadcrumb is swiped from app/repo_tree.go.
function SnippetToBreadcrumb(repo, rev, path, startLine, endLine, defURL, cb) {
	path = path[0] === "/" ? path.substring(1) : path;

	var curPath = router.fileURL(repo, rev, "");
	var segs = path.split("/");
	var breadcrumb = [];

	var onSnippetClick = function onSnippetClick(file, start, end, url, evt) {
		if (typeof cb === "function") {
			cb(file, start, end, url);
			evt.preventDefault();
		}
	};

	for (var i = 0; i < segs.length; i++) {
		if (i > 0) breadcrumb.push(<span key={`ex_sep_${i}`} className="sep">/</span>);
		if (segs[i] === ".") break;

		var linktext = segs[i];
		if (i === segs.length - 1 && startLine !== 0) {
			linktext += endLine !== 0 ? `:${startLine}-${endLine}` : `:${startLine}`;
			var href = `${router.fileURL(repo, rev, path)}?startline=${startLine}&endline=${endLine}&defUrl=${defURL}`;

			breadcrumb.push(
				<a key={repo+rev+path+defURL+linktext}
					href={href}
					target="_blank"
					onClick={onSnippetClick.bind(this, {
						Path: path,
						RepoRev: {
							URI: repo,
							Rev: rev,
						},
					}, startLine, endLine, defURL)}>
					{linktext}
				</a>
			);
		} else {
			breadcrumb.push(<a key={curPath+segs[i]+linktext} href={curPath + segs[i]}>{linktext}</a>);
		}
		curPath += `${segs[i]}/`;
	}

	return breadcrumb;
}
