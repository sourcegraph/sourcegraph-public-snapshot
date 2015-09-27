var React = require("react");
var classNames = require("classnames");
var DiffStatScale = require("./DiffStatScale");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");

var DiffFileList = React.createClass({

	propTypes: {
		// Function triggered when a file is clicked in the list. It receives as
		// parameters: FileDiff (Backbone.Model) and Event.
		onFileClick: React.PropTypes.func,

		// urlBase is any possible prefix to be added to the URL that should open
		// the files in the list. The file name will be appended to this URL.
		urlBase: React.PropTypes.string,
	},

	mixins: [ModelPropWatcherMixin],

	getDefaultProps() {
		return {
			onFileClick() {},
			urlBase: "",
		};
	},

	/*
	 * @description Triggered when the header of the file list is clicked.
	 * @private
	 */
	_onHeaderClick() {
		this.setState({closed: !this.state.closed});
	},

	render() {
		var fileListCx = classNames({
			"file-list": true,
			"closed": this.state.closed === true,
		});

		return this.state.models.length ? (
			<div className={fileListCx}>
				<a className="file-list-header" onClick={this._onHeaderClick}>
					<i className={this.state.closed ? "fa fa-icon fa-plus-square-o" : "fa fa-icon fa-minus-square-o"} />
					<b>Files</b> <span className="count">( {this.state.models.length} )</span>
					<span className="pull-right stats">
						<span className="additions-color">+{this.props.stats.Added}</span>
						<span className="deletions-color">-{this.props.stats.Deleted}</span>
					</span>
				</a>

				<ul className="file-list-items">
					{this.state.models.map(fd => {
						var baseName = fd.getBaseFilename(),
							headName = fd.getHeadFilename(),
							stats = fd.get("Stats"), drafts = 0, comments = 0;

						if (this.props.reviews) {
							if (headName) {
								drafts += this.props.reviews.drafts.where({Filename: headName}).length;
								comments += this.props.reviews.comments.where({Filename: headName}).length;
							}
							if (baseName && baseName !== headName) {
								drafts += this.props.reviews.drafts.where({Filename: baseName}).length;
								comments += this.props.reviews.comments.where({Filename: baseName}).length;
							}
						}

						return (
							<li key={`fname-${fd.cid}`} className="file-list-item">
								<a href={`${this.props.urlBase}/${fd.getHeadFilename() || fd.getBaseFilename()}`} onClick={this.props.onFileClick.bind(this, fd)}>
									{!baseName ? <i className="fa change-type octicon octicon-diff-added additions-color" /> : null}
									{!headName ? <i className="fa change-type octicon octicon-diff-removed deletions-color" /> : null}
									{Boolean(baseName) && Boolean(headName) ? <i className="fa change-type octicon octicon-diff-modified changes-color" /> : null}

									{baseName && headName && baseName !== headName ? (
										<span>{baseName} <i className="fa fa-icon fa-long-arrow-right" />&nbsp;</span>
									) : null}

									{headName || baseName}

									{drafts > 0 ? (
										<span className="draft-count">
											<i className="fa fa-comment-o"></i> {drafts}
										</span>
									) : null}

									{comments > 0 ? (
										<span className="comment-count">
											<i className="fa fa-comment"></i> {comments}
										</span>
									) : null}

									<span className="pull-right stats">
										<span className="additions-color">+{stats.Added}</span>
										<span className="deletions-color">-{stats.Deleted}</span>
										<DiffStatScale Stat={stats} />
									</span>
								</a>
							</li>
						);
					})}
				</ul>
			</div>
		) : null;
	},
});

module.exports = DiffFileList;
