var React = require("react");
var Draggable = require("react-draggable");
var globals = require("../globals");
var router = require("../routing/router");
var classNames = require("classnames");

var CodeFileActions = require("../actions/CodeFileActions");
var TokenPopupStore = require("../stores/TokenPopupStore");
var ExamplesView = require("./ExamplesView");

var DiscussionList = require("./DiscussionList");
var DiscussionSnippet = require("./DiscussionSnippet");
var DiscussionCreateForm = require("./DiscussionCreateForm");
var DiscussionView = require("./DiscussionView");

/**
 * @description Manages the display of the window which shows definition
 * documentation and usage examples
 */
var TokenPopupView = React.createClass({

	propTypes: {
		// Indicates that data is loading.
		loading: React.PropTypes.bool,
	},

	getInitialState() {
		return TokenPopupStore.attributes;
	},

	componentDidMount() {
		TokenPopupStore.on("change", this._change);
		window.addEventListener("keyup", this._keyPress);
	},

	componentWillUnmount() {
		TokenPopupStore.off("change", this._change);
		window.removeEventListener("keyup", this._keyPress);
	},

	/**
	 * @description Synchronizes the components state with the store. This
	 * method is called whenever the store 'change's.
	 * @returns {void}
	 * @private
	 */
	_change() {
		this.setState(TokenPopupStore.attributes);
	},

	/**
	 * @description Called when token is clicked inside example's code view.
	 * @param {CodeTokenModel} token - Model of clicked token.
	 * @param {Event} e - Click event.
	 * @returns {void}
	 * @private
	 */
	_clickExampleToken(token, e) {
		CodeFileActions.selectToken(token, {silent: true}); // does not register with router
	},

	_getDoc() {
		var doc = {
			header: [
				<h1 className="qualified-name"
					key="headerName"
					// This is OK because QualifiedName is guaranteeed to be
					// sanitized in ui/def.go by serveDef.
					dangerouslySetInnerHTML={this.state.QualifiedName} />],

			body: null,
		};

		if (this.state.Found === false) {
			doc.body = (
				<section key="missingMessage" className="doc">
					Definition of{" "}
					// This is OK because QualifiedName is guaranteeed to be
					// sanitized in ui/def.go by serveDef.
					<span className="qualified-name" dangerouslySetInnerHTML={this.state.QualifiedName} />
					{" "}is not available.
				</section>
			);
			return doc;
		}

		if (this.state.extraDefinitions.length > 0) {
			doc.header.push(
				<ul key={this.state.URL} className="overlapping-defs ui-tree">
					{this.state.extraDefinitions.map(def =>
						<li onClick={e => this._clickAdjacentDefinition(def, e)}>
							<i className="fa fa-tag" />
							<span className="qualified-name"
								// This is OK because QualifiedName is guaranteeed to
								// be sanitized in ui/def.go by serveDef.
								dangerouslySetInnerHTML={def.QualifiedName} />
						</li>
					)}
				</ul>
			);
		}

		if (this.state.Data && this.state.Data.DocHTML) {
			// This is also OK because DocHTML is sanitized by the app (not
			// untrusted federation root server) where the Def (Data) comes
			// from. This happens in util/handlerutil/repo.go by
			// GetDefCommon.
			doc.body = <section className="doc" dangerouslySetInnerHTML={this.state.Data.DocHTML} />;
		}

		return doc;
	},

	/**
	 * @description Called when one of the additional overlapping definitions is clicked in the
	 * pop-up (occurs in Scala).
	 * @param {object} def - Definition.
	 * @param {Event} e - Click event.
	 * @returns {void}
	 * @private
	 */
	_clickAdjacentDefinition(def, e) {
		CodeFileActions.selectAlternativeDefinition(router.defURL(def.Data));
		e.preventDefault();
	},

	/**
	 * @description Called when the "Go to definition" button is clicked.
	 * @param {event} e - Click event.
	 * @returns {void}
	 * @private
	 */
	_clickGoToDefinition(e) {
		if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) {
			return; // allow open in new tab
		}
		CodeFileActions.navigateToDefinition(TokenPopupStore.attributes);
		e.preventDefault();
	},

	/**
	 * @description Called during the keyup event bound in componentDidMount.
	 * @param {event} e - Keyup event.
	 * @returns {void}
	 * @private
	 */
	_keyPress(e) {
		switch (e.keyCode) {
		case 27:
			// Close on escape key press.
			this._close(e);
			break;
		}
	},

	/**
	 * @description Closes the token popup.
	 * @param {event} e - Click/keyup event.
	 * @returns {void}
	 * @private
	 */
	_close(e) {
		CodeFileActions.deselectTokens();
		this.setState({closed: true});
		e.preventDefault();
	},

	/**
	 * @description Returns true if the active definition has discussions.
	 * @returns {bool} True if the active definition has discussion.
	 * @private
	 */
	_hasDiscussions() {
		return this.state.topDiscussions.models.length > 0;
	},

	/**
	 * @description Returns an array of JSX where each pair of arguments
	 * defines the final form of the array. In each pair, the first argument
	 * is a boolean. If it resolves to true, the second argument in the pair
	 * will be part of the final array. For example:
	 *
	 * 	// returns [<Button />, <Dropdown />]
	 * 	this._renderConditional(
	 * 		true, <Button />,
	 * 		true, <Dropdown />,
	 * 		false, <Input />
	 * 	);
	 *
	 * @returns {Array<jsx>} Array of React.JSX
	 * @private
	 */
	_renderConditional() {
		var args = Reflect.apply(Array.prototype.slice, arguments);
		if (args.length%2 !== 0) {
			console.error("_renderConditional method received odd number of args.");
			return null;
		}
		var jsx = [];
		for (var i = 0; i < args.length-1; i += 2) {
			if (args[i] === true) jsx.push(args[i+1]);
		}
		return jsx;
	},


	/**
	 * @description Generates the documentation header component based on the information
	 * available in the state.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderDocumentationHeader() {
		var hasMoreDefinitions = this.state.extraDefinitions.length > 0;

		return this._renderConditional(
			true, (
				<h1 className="qualified-name"
					key="headerName"
					dangerouslySetInnerHTML={this.state.QualifiedName} />
			),
			hasMoreDefinitions, (
					<ul key={this.state.URL} className="overlapping-defs ui-tree">
						{this.state.extraDefinitions.map(def =>
							<li onClick={e => this._clickAdjacentDefinition(def, e)}>
								<i className="fa fa-tag" />
								<span className="qualified-name"
									dangerouslySetInnerHTML={def.QualifiedName} />
							</li>
						)}
					</ul>
			)
		);
	},

	/**
	 * @description Generates the documentation body component based on the information
	 * available in the state.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderDocumentationBody() {
		if (this.state.Found === false) {
			// Definition is correct but wasn't indexed.
			return (
				<section key="missingMessage" className="doc">
					Definition of{" "}
					<span className="qualified-name" dangerouslySetInnerHTML={this.state.QualifiedName} />
					{" "}is not available.
				</section>
			);
		}
		// Attach body, if non-empty.
		if (this.state.Data) {
			var doc = this.state.Data.DocHTML || "";
			if (doc.__html && doc.__html !== "") {
				return <section className="doc" dangerouslySetInnerHTML={this.state.Data.DocHTML} />;
			}
		}

		return null;
	},

	/**
	 * @description Creates the toolbar based on state.
	 * @returns {jsx} The toolbar JSX mark-up.
	 * @private
	 */
	_renderToolbar() {
		var isNotDefaultPage = this.state.page.type !== globals.PopupPages.DEFAULT;
		var hasValidDefinition = this.state.URL && this.state.Found;
		var isDefinition = this.state.type === globals.TokenType.DEF;

		return this._renderConditional(
			isNotDefaultPage, (
				<a key="back-to-main" className="btn btn-toolbar btn-default" onClick={CodeFileActions.showPopupPageDefault}>
					<span className="octicon octicon-arrow-left" /> Back to token
				</a>
			),
			hasValidDefinition, (
				<a key="popup-definition-btn" className="btn btn-toolbar btn-default" href={this.state.URL} onClick={this._clickGoToDefinition}>
					Go to definition
				</a>

			),
			isDefinition, (
				<a key="popup-embed-btn" className="btn btn-toolbar btn-default" href={`${this.state.URL}/.share`}>
					Embed
				</a>
			)
		);
	},

	/**
	 * @description Returns the JSX for the "Top rated discussions" snippet.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderTopDiscussionsSnippet() {
		var page = this.state.page;
		var url = this.state.URL;

		// Can't render discussions if we don't have a selected unit
		if (!url) {
			return null;
		}

		return (
			<DiscussionSnippet key={new Date()}
				model={this.state.topDiscussions}
				toolbar={page.type === globals.PopupPages.DEFAULT}
				defKey={url}
				onCreate={CodeFileActions.createDiscussion}
				onList={CodeFileActions.showPopupPageDiscussionList}
				onClick={CodeFileActions.openDiscussion} />
		);
	},

	/**
	 * @description Renders the pop-up page to list discussions.
	 * @param {object} data - Discussion list.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderDiscussionListPage(data) {
		return (
			<DiscussionList
				model={data}
				defName={this.state.QualifiedName}
				defKey={this.state.URL}
				onCreate={CodeFileActions.createDiscussion}
				onClick={CodeFileActions.openDiscussion} />
		);
	},

	/**
	 * @description Renders the pop-up page to create a discussion.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderCreateDiscussionPage() {
		var name = this.state.QualifiedName;
		var key = `create${name}`;

		return this._renderConditional(
			true, (
				<DiscussionCreateForm key={key}
					defName={name}
					onCreate={CodeFileActions.submitDiscussion}
					onCancel={CodeFileActions.showPopupPageDefault} />
			),

			this._hasDiscussions(), this._renderTopDiscussionsSnippet()
		);
	},

	/**
	 * @description Renders the pop-up page that shows a discussion.
	 * @param {object} data - Discussion.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderDiscussionPage(data) {
		return (
			<DiscussionView
				defName={this.state.QualifiedName}
				defKey={this.state.URL}
				onList={CodeFileActions.showPopupPageDiscussionList}
				onCreate={CodeFileActions.createDiscussion}
				onComment={CodeFileActions.submitDiscussionComment}
				model={data} />
		);
	},

	/**
	 * @description Renders the default pop-up page.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderDefaultPage() {
		var components = [
			(<div className="token-view" key={this.state.URL}>
				<section className="docHTML">
					<div className="header">
						{this._renderDocumentationHeader()}
					</div>
					{this._renderDocumentationBody()}
				</section>

				<ExamplesView
					{...this.props}
					onTokenFocus={CodeFileActions.focusToken}
					onTokenBlur={CodeFileActions.blurTokens}
					onTokenClick={this._clickExampleToken}
					onChangePage={CodeFileActions.selectExample}
					onShowSnippet={CodeFileActions.changeState}
					def={this.state.URL}
					model={this.state.examplesModel} />
			</div>),
		];
		if (globals.Features.Discussions) {
			components.push(this._renderTopDiscussionsSnippet());
		}
		return components;
	},

	/**
	 * @description Renders the jsx for the current pop-up page based on
	 * state.
	 * @returns {jsx} React.JSX
	 * @private
	 */
	_renderActivePage() {
		var page = this.state.page;
		switch (page.type) {
		case globals.PopupPages.LIST_DISCUSSION:
			return this._renderDiscussionListPage(page.data);

		case globals.PopupPages.NEW_DISCUSSION:
			return this._renderCreateDiscussionPage();

		case globals.PopupPages.VIEW_DISCUSSION:
			return this._renderDiscussionPage(page.data);

		default:
			return this._renderDefaultPage();
		}
	},

	render() {
		var classes = classNames({
			"token-details": true,
			"error": this.state.error,
			"closed": this.state.closed,
		});

		var toolbar = this._renderToolbar();
		var content = this._renderActivePage();

		return (
			<Draggable handle="header.toolbar">
				<div className={classes}>
					<div className="body">
						<header className="toolbar">
							{toolbar}
							<a className="close top-action" onClick={this._close}>×</a>
						</header>
						{content}
					</div>
				</div>
			</Draggable>
		);
	},
});

module.exports = TokenPopupView;
