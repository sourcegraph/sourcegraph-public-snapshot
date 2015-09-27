var React = require("react");
var RepoRevSwitcher = require("./RepoRevSwitcher");
var RepoBuildIndicator = require("./RepoBuildIndicator");
var router = require("../routing/router");
var classNames = require("classnames");

var CodeFileToolbarView = React.createClass({

	propTypes: {
		// Contains information about the current file, such as Path and RepoRev.
		file: React.PropTypes.object,

		// Indicates whether a loading event is occurring on the container.
		loading: React.PropTypes.bool,

		// The number of references in the current file.
		numRefs: React.PropTypes.number,

		// If true, indicates that the maximum number of references has been
		// reached and the file is only partially linked.
		maxRefs: React.PropTypes.bool,
	},

	getInitialState() {
		return {
			loading: this.props.loading || false,
			file: this.props.file || null,
		};
	},

	/**
	 * @description Creates and links the breadcrumb data for the toolbar, given the current file.
	 * @returns {JSX} - Breadcrumb template
	 * @private
	 */
	_getBreadcrumb() {
		var rev = this.props.file.RepoRev,
			basePath = `/${rev.URI}@${rev.Rev||rev.CommitID}/.tree`,
			uriSegs = rev.URI.split("/"),
			breadcrumb = [<a key="feature_toolbar_breadcrumb" href={basePath}>{uriSegs[uriSegs.length-1]}</a>];

		this.props.file.Path.split("/").forEach((seg, i) => {
			basePath += `/${seg}`;

			breadcrumb.push(
				<span key={`separator${i.toString()}`}> / </span>,
				<a key={basePath+seg} href={basePath}>{seg}</a>
			);
		});

		return breadcrumb;
	},

	render() {
		if (!this.props.file) return null;

		var breadcrumb = this._getBreadcrumb(),
			rev = this.props.file.RepoRev;

		var fileIconClasses = classNames({
			"fa": true,
			"fa-file": !this.props.loading,
			"fa-spinner fa-spin": this.props.loading,
		});

		var embedLink = `/${rev.URI}@${rev.Rev||rev.CommitID}/.tree/${this.props.file.Path}/.share`,
			snippet = this.props.snippet;

		if (snippet && snippet.start && snippet.end) {
			embedLink += `?StartLine=${snippet.start}&EndLine=${snippet.end}`;
		}

		var buildInfo = null;
		if (this.props.buildInfo && this.props.buildInfo.CommitsBehind) {
			var latestFileURL = router.fileURL(this.props.file.RepoRev.URI, this.props.latestCommit.ID, this.props.file.Path);
			buildInfo = (<i className="build-info">
				Code from {this.props.buildInfo.CommitsBehind} commits behind <a href={latestFileURL}>head</a> is shown
			</i>);
		} else if (this.props.numRefs) {
			buildInfo = <i className="build-info">{this.props.numRefs} linked</i>;
		} else {
			buildInfo = <i className="build-info">No links</i>;
		}

		return (
			<div className="code-file-toolbar">
				<div className="file">
					<i className={fileIconClasses} />{breadcrumb}

					<RepoBuildIndicator
						RepoURI={rev.URI}
						Rev={rev.CommitID}
						btnSize="btn-xs"
						tooltipPosition="bottom" />

					{buildInfo}
				</div>

				<div className="actions">
					<RepoRevSwitcher repoSpec={rev.URI}
						rev={rev.Rev||rev.CommitID}
						path={this.props.file.Path}
						alignRight={true} />

					<a className="share top-action btn btn-default btn-xs"
						aria-hidden="true"
						href={embedLink}
						data-tooltip={true} title="Select text to specify a line range">Embed</a>
				</div>
			</div>
		);
	},
});

module.exports = CodeFileToolbarView;
