var React = require("react");
var Draggable = require("react-draggable");
var ExamplesView = require("./ExamplesView");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var DiffActions = require("../actions/DiffActions");
var classNames = require("classnames");

var DiffPopup = React.createClass({

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

		// This is also OK because DocHTML is sanitized by the app (not
		// untrusted federation root server) where the Def (Data) comes
		// from. This happens in util/handlerutil/repo.go by GetDefCommon.
		doc.body = <section className="doc" dangerouslySetInnerHTML={this.state.Data.DocHTML} />;
		return doc;
	},

	_close(e) {
		this.setState({closed: true});
		if (typeof this.props.onClose === "function") {
			this.props.onClose(e);
			e.preventDefault();
		}
	},

	render() {
		var classes = classNames({
			"token-details": true,
			"error": this.state.error,
			"closed": this.state.closed,
		});

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
									&nbsp;in <i>{this.state.File.Path}</i> @ {this.state.File.RepoRev.CommitID.substring(0, 7)}
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
							onTokenFocus={DiffActions.focusToken}
							onTokenBlur={DiffActions.blurTokens.bind(this, undefined)}
							onTokenClick={DiffActions.selectToken}
							onChangePage={DiffActions.selectExample}
							def={this.state.URL}
							model={this.state.examplesModel} />
					</div>
				</div>
			</Draggable>
		);
	},
});

module.exports = DiffPopup;
