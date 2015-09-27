var React = require("react");
var $ = require("jquery");
var notify = require("../components/notify");

var CodeFileActions = require("../actions/CodeFileActions");
var CodeFileStore = require("../stores/CodeFileStore");
var CodeFileRouter = require("../routing/CodeFileRouter");

var CodeFileToolbarView = require("./CodeFileToolbarView");
var CodeView = require("./CodeView");
var TokenPopupView = require("./TokenPopupView");
var TokenPopoverView = require("./TokenPopoverView");
var ContextMenu = require("./ContextMenu");

/**
 * @description CodeFileView displays the contents of a file set through its
 * source property, along with a toolbar, a details box and hover events.
 */
var CodeFileView = React.createClass({

	propTypes: {
		// The URL source of the file to be displayed.
		source: React.PropTypes.string,

		// Any valid server response object. This is populated when the file view
		// is pre-loaded into the document on the server-side.
		data: React.PropTypes.object,
	},

	getInitialState() {
		this.codeViewEventBound = false;
		return CodeFileStore.attributes;
	},

	componentDidMount() {
		CodeFileRouter.start();

		if (this.isMounted()) this._bindEvents();

		if (this.props.data !== null && typeof this.props.data === "object") {
			CodeFileActions.renderPreloaded(this.props.data);
		} else if (this.props.source) {
			CodeFileActions.selectFile(this.props.source);
		}
	},

	componentDidUpdate(prevProps, prevState) {
		if (this.isMounted() && !this.codeViewEventBound) {
			$(this.getDOMNode())
				.find(".line-numbered-code")
				.on("click", CodeFileActions.focusCodeView);

			this.codeViewEventBound = true;
		}

		CodeFileRouter.matchInitialHashState();
		if (this.state.file && this.state.file.Path !== prevState.file.Path) {
			notify.info(`Loaded file <i>${this.state.file.Path}</i>...`);
		}
	},

	componentWillUnmount() {
		CodeFileStore.off("change");
		CodeFileStore.off("scrollTop");
	},

	codeViewEventBound: null,

	/**
	 * @description Binds to events on the store.
	 * @returns {void}
	 * @private
	 */
	_bindEvents() {
		CodeFileStore.on("change", () => this.setState(CodeFileStore.attributes));
		CodeFileStore.on("scrollTop", this.refs.codeView.scrollTo);
	},

	_onContextMenuClick(item, evt) {
		CodeFileActions.selectAlternativeDefinition(item.URL);
	},

	render() {
		var features = [
			<TokenPopupView key="feature_popup" />,

			<TokenPopoverView key="feature_popover" model={this.state.popoverModel} />,

			<CodeFileToolbarView key="feature_toolbar"
				file={this.state.file}
				snippet={this.state.snippet}
				loading={this.state.loading}
				numRefs={this.state.numRefs}
				maxRefs={this.state.maxRefs}
				buildInfo={this.state.buildInfo}
				latestCommit={this.state.latestCommit} />,

			<ContextMenu key="context_menu"
				model={this.state.contextMenuModel}
				onClick={this._onContextMenuClick} />,
		];

		return (
			<div className="code-view-react">
				{features}
				<CodeView model={this.state.codeModel}
					loading={this.state.loading}
					lineNumbers={true}
					onTokenClick={CodeFileActions.selectToken}
					onTokenFocus={CodeFileActions.focusToken}
					onTokenBlur={CodeFileActions.blurTokens}
					ref="codeView"
					tilingAfter={1000} />
			</div>
		);
	},
});

module.exports = CodeFileView;
