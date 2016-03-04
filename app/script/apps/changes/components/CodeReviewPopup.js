var React = require("react");
var Draggable = require("react-draggable");
var classNames = require("classnames");
var ExamplesView = require("../../../components/ExamplesView");
var ModelPropWatcherMixin = require("../../../components/mixins/ModelPropWatcherMixin");

var CodeReviewPopup = React.createClass({

	mixins: [ModelPropWatcherMixin],

	_getDoc() {
		var doc = {
			header: [
				<h1 className="qualified-name"
					key={this.state.URL || "undefined"}
					// This is OK because QualifiedName is guaranteeed to be
					// sanitized in ui/def.go by serveDef.
					dangerouslySetInnerHTML={this.state.QualifiedName} />],

			body: null,
		};

		if (!this.state.Data) {
			return doc;
		}
		// This is OK because DocHTML is sanitized by the app (not
		// untrusted federation root server) where the Def (Data) comes
		// from. This happens in util/handlerutil/repo.go by GetDefCommon.
		doc.body = <section className="doc" dangerouslySetInnerHTML={this.state.Data.DocHTML} />;
		return doc;
	},

	_close(e) {
		if (typeof this.props.onClose === "function") {
			this.props.onClose(e);
			e.preventDefault();
		}
	},

	render() {
		var classes = classNames({
			"token-details": true,
			"token-details-review": true,
			"error": this.state.error,
			"closed": this.state.closed,
		});

		// In this case, there was an error fetching the definition. This could be
		// due to a variety of different reasons. Avoid causing an exception below
		// which would break all future CodeReviewPopups.
		if (this.state.File && !this.state.File.RepoRev.CommitID) {
			console.error("CodeReviewPopup error: File.RepoRev.CommitID ==", this.state.File.RepoRev.CommitID);
			console.error("CodeReviewPopup note: File ==", this.state.File);
			return null;
		}

		var doc = this._getDoc();

		return (
			<Draggable handle="header.toolbar">
				<div className={classes}>
					<div className="body">
						<header className="toolbar">
							<a className="btn btn-toolbar btn-default" target="_blank" href={this.state.URL}>
								Open definition
							</a>
								{this.state.File ? <span className="fileInfo">
									&nbsp;in <i className="file-path" title={this.state.File.Path}>{this.state.File.Path}</i> @ {this.state.File.RepoRev.CommitID.substring(0, 7)}
								</span> : null}
							<a className="close top-action" onClick={this._close}>×</a>
						</header>

						<section className="docHTML">
							<div className="header">
								{doc.header}
							</div>
							{doc.body}
						</section>

						<ExamplesView
							{...this.props}
							def={this.state.URL}
							model={this.state.examplesModel} />
					</div>
				</div>
			</Draggable>
		);
	},
});

module.exports = CodeReviewPopup;
