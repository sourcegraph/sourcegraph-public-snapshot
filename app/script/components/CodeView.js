var React = require("react");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var CodeLineView = require("./CodeLineView");
var CodeModel = require("../stores/models/CodeModel");
var classNames = require("classnames");
var $ = require("jquery");
var debounce = require("../debounce");

/**
 * @description CodeView holds a collection of lines and tokens received from the server.
 * To initialize the CodeView, the model property needs to be set to a valid CodeModel.
 *
 * CodeView may render 2 types of views based on its props: a tiled view and the regular
 * view.
 *
 * The tiled view will be rendered if the "tilingAfter" property is set and the number of
 * lines of code is larger than its value. In this case, the code view will be as large
 * as its size in CSS and will take control of its own scrolling. This view uses a
 * tiling technique that significantly improves initial loading time, similar to:
 * http://blog.atom.io/2015/06/24/rendering-improvements.html ("Tiles to the rescue!")
 * Initial loading time on any file that has more than "tilingAfter" lines is 4-6 seconds.
 * Without this technique, on a file with 14k lines the loading time is 26s.
 *
 * The regular view will render a regular table and will leave scroll control up to the
 * browser, or any chosen container.
 */
var CodeView = React.createClass({

	propTypes: {
		// The model that the CodeView should render
		model: React.PropTypes.instanceOf(CodeModel).isRequired,

		// Optionally pass the loading property, to disable the CodeView
		// when loading occurs.
		loading: React.PropTypes.bool,

		// If the property is exclusively set to false, line numbers will be
		// hidden.
		lineNumbers: React.PropTypes.bool,

		// The function to be called on click. It will receive as arguments the
		// CodeTokenModel that was click, and the click event. Default will be
		// automatically prevented.
		onTokenClick: React.PropTypes.func,

		// The function to be called on 'mouseenter'. It will receive as arguments the
		// CodeTokenModel, and the event. Default will be automatically prevented.
		onTokenFocus: React.PropTypes.func,

		// The function to be called on 'mouseleave'. It will receive as arguments the
		// CodeTokenModel, and the event. Default will be automatically prevented.
		onTokenBlur: React.PropTypes.func,

		// When set, if the number of lines in the code view goes past this number,
		// the code view enters tile mode. Tile mode currently only works
		// by using it's own scroll container - an external container (such
		// as the browser) has no control in this mode.
		tilingAfter: React.PropTypes.number,
	},

	mixins: [ModelPropWatcherMixin],

	getInitialState() {
		return {tileInView: 1, lineHeight: 19};
	},

	componentDidUpdate() {
		var node = $(this.getDOMNode());

		node.off("scroll");
		if (this._isTiled()) {
			node.on("scroll", debounce(this._onScroll, 60));
		}
	},

	componentWillUnmount() {
		$(this.getDOMNode()).off("scroll");
	},

	/**
	 * @description Scrolls the passed line or token into view by changing the
	 * browser's scroll position, or, in case of a tiled view, it's own scroll
	 * container.
	 * @param {CodeLineModel|CodeTokenModel} lineOrToken - Line or token
	 * @returns {void}
	 */
	scrollTo(lineOrToken) {
		var amount;
		var node = $("html, body");
		var duration = 400;

		if (this._isTiled()) {
			var lineNumber = lineOrToken.get("number") || lineOrToken.get("line").get("number");
			amount = this.state.lineHeight * lineNumber - 80;
			node = $(this.getDOMNode());
			duration = 0;
		} else {
			amount = lineOrToken.getRelativePosition().top - 200;
		}

		node.animate({scrollTop: amount}, duration, "linear");
	},

	/**
	 * @description The callback that is triggered when the user scrolls inside
	 * the code view. This callback is only bound in tile view. In all other
	 * cases, the browser takes control of browsing.
	 * @returns {void}
	 */
	_onScroll() {
		var linesPerTile = Math.ceil(screen.height / this.state.lineHeight);
		var tile = Math.round($(this.getDOMNode()).scrollTop() / (linesPerTile * this.state.lineHeight));

		if (tile !== this.state.tileInView) {
			this.setState({tileInView: tile});
		}
	},

	/**
	 * @description Tells us if the view needs to be, or already is tiled.
	 * @returns {bool} Will be true if the view is tiled.
	 */
	_isTiled() {
		var threshold = this.props.tilingAfter;
		return threshold && this.state.lines.length > threshold;
	},

	/**
	 * @description Returns the tile at the specified index based on current
	 * scroll position. It can either be a tile with lines of code or an
	 * off-screen placeholder.
	 * @param {number} i - Index of tile to generate.
	 * @param {number} linesPerTile - Number of lines in one tile.
	 * @param {number} totalTiles - Total number of tiles.
	 * @param {string} cx - Extra CSS classes to be placed on the tile.
	 * @returns {jsx} The JSX of the tile.
	 */
	_makeTile(i, linesPerTile, totalTiles, cx) {
		var tileInView = this.state.tileInView;

		// Is this tile out of view? If so, return a placeholder.
		if (i <= tileInView - 2 || i >= tileInView + 3) {
			return (
				<div key={`placeholder-${i}`}
					className={"line-numbered-code placeholder tile"}
					style={{height: linesPerTile * this.state.lineHeight}} />
			);
		}

		// This is either the tile in view, the one above or the one below it.
		var startLine = i * linesPerTile;
		var endLine = i < totalTiles-1 ? (i+1)*linesPerTile : this.state.lines.length;
		var tileHeight = i === totalTiles-1 ? "auto" : linesPerTile * this.state.lineHeight;

		return (
			<table key={`tile-${i}`} className={`${cx} tile`} style={{height: tileHeight}}>
				<tbody>
					{this.state.lines.slice(startLine, endLine).map(
						line => <CodeLineView {...this.props} key={line.cid} model={line} style={{height: this.state.lineHeight}} />
					)}
				</tbody>
			</table>
		);
	},

	/**
	 * @description Renders the tiled view. The tiled view is rendered when the
	 * threshold in the "tilingAfter" property is set and the number of lines of code
	 * is larger than its value.
	 * @param {string} cx - Classes to be set on tiles.
	 * @returns {jsx} The rendered view.
	 */
	_renderTiles(cx) {
		var linesPerTile = Math.ceil(screen.height / this.state.lineHeight);
		var totalTiles = Math.ceil(this.state.lines.length / linesPerTile);
		var tiles = [];

		for (var i = 0; i < totalTiles; i++) {
			tiles.push(this._makeTile(i, linesPerTile, totalTiles, cx));
		}

		return <div className="tile-view scroll-container">{tiles}</div>;
	},

	render() {
		if (!this.state.lines.length) return <i className="file-loader fa fa-spinner fa-spin" />;

		var cx = classNames({
			"pale": this.props.loading,
			"theme-default": this.props.theme === "default",
			"line-numbered-code": true,
		});

		return this._isTiled() ? this._renderTiles(cx) : (
			<table className={cx}>
				<tbody>
					{this.state.lines.map(
						line => <CodeLineView {...this.props} key={line.cid} model={line} />
					)}
				</tbody>
			</table>
		);
	},
});

module.exports = CodeView;
