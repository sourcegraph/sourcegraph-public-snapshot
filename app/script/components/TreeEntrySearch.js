var React = require("react");
var client = require("../client");
var $ = require("jquery");
var router = require("../routing/router");
var Fuze = require("fuse.js");
var classNames = require("classnames");
var notify = require("./notify");

var FILE_LIMIT = 15;

var TreeEntrySearch = React.createClass({
	getInitialState() {
		return {
			visible: false,
			loading: false,
			set: [],
			selectionIndex: 0,
		};
	},

	componentDidMount() {
		$(document).keyup(e => {
			if (!this.isMounted() || $(e.target).is("input, select, textarea") && e.keyCode !== 27) {
				return;
			}
			switch (e.keyCode) {
			case 84: this._focusInput(); break;
			case 27: this._blurInput(); break;
			default: // noop
			}
		});
	},

	componentWillUnmount() {
		$(document).off("keyup");
	},

	/**
	 * @description FuzzySet contains the set of files through which
	 * fuzzy searching will be applied.
	 * @type {Fuze}
	 */
	fuzzySet: null,

	/**
	 * @description Holds the set of all files.
	 * @type {Array}
	 */
	files: null,

	/**
	 * @description Holds the timeout that triggers the search. This is to avoid
	 * repetitive searching while typing.
	 * @type {Object}
	 */
	searchTimeout: null,

	/**
	 * @description Triggered when the search is shown. If the files for this
	 * repository have not yet been loaded, it loads them.
	 * @returns {void}
	 * @private
	 */
	_focusInput() {
		if ($("body").data("file-search-disabled")) {
			return null;
		}

		this.setState({
			visible: true,
			loading: true,
			selectionIndex: 0,
		}, () => $(".tree-search-input input").focus());

		// If we already loaded the files once for this repo, don't do it again.
		if (this.fuzzySet && this.fuzzySet.list.length > 0) {
			return this.setState({loading: false});
		}

		client.repoFiles(this.props.repo, this.props.rev).then(
			this._loadedList, () => notify.error("Couldn't load file list")
		);
	},

	/**
	 * @description Triggered after the list of files for this repository has been
	 * returned from the server.
	 * @param {Array<string>} list - List of files.
	 * @returns {void}
	 * @private
	 */
	_loadedList(list) {
		this.fuzzySet = new Fuze(list, {
			distance: 1000,
			location: 0,
			threshold: 0.1,
		});

		this.files = this.fuzzySet.list;

		this.setState({set: this.files}, () => this.setState({loading: false}));
	},

	/**
	 * @description Triggered when the input is hidden by pressing Escape or clicking
	 * on the overlay.
	 * @returns {void}
	 * @private
	 */
	_blurInput() {
		$(".tree-search-input input").val("").blur();

		this.setState({
			visible: false,
			loading: false,
		});
	},

	/**
	 * @description Triggered when typing in the input. Based on the key pressed,
	 * this may either perform a search or navigate the result list.
	 * @param {Event} e - Event
	 * @returns {void}
	 * @private
	 */
	_onType(e) {
		if (!this.isMounted()) return;

		var idx, set, max;
		switch (e.key) {
		case "ArrowDown":
			idx = this.state.selectionIndex;
			set = this._getSet();
			max = set.length > FILE_LIMIT ? FILE_LIMIT : set.length;

			this.setState({
				selectionIndex: idx + 1 >= max ? 0 : idx + 1,
			});

			e.preventDefault();
			break;

		case "ArrowUp":
			idx = this.state.selectionIndex;
			set = this._getSet();
			max = set.length > FILE_LIMIT ? FILE_LIMIT : set.length;

			this.setState({
				selectionIndex: idx < 1 ? max-1 : idx-1,
			});

			e.preventDefault();
			break;

		case "Enter":
			var fileURL = router.fileURL(this.props.repo, this.props.rev, this._getSet()[this.state.selectionIndex]);
			window.location = fileURL;
			e.preventDefault();
			break;

		default:
			this._performSearch();
		}
	},

	_performSearch() {
		if (this.fuzzySet === null) {
			return // file set is not loaded yet.
		}
		var term = $(".tree-search-input input").val();
		if (this.searchTimeout === null && term !== "") {
			this.searchTimeout = setTimeout(() => {
				this.setState({
					set: term === "" ? this.files : this._search(term),
					selectionIndex: 0,
				});

				this.searchTimeout = null;
			}, 500);
		}
	},

	/**
	 * @description Performs a search in the fuzzy set.
	 * @param {string} term - Term to filter by.
	 * @returns {Array} - Array of filtered files that match term.
	 * @private
	 */
	_search(term) {
		return this.fuzzySet.search(term).map(i => this.files[i]);
	},

	/**
	 * @description Returns the set of files that the user currently sees. This
	 * set may either be the unfiltered array of files, or the subset resulting
	 * from the search term.
	 * @returns {Array} - Array of all files.
	 * @private
	 */
	_getSet() {
		var term = $(".tree-search-input input").val();
		return (this.state.set.length || term !== "") ? this.state.set : this.files || [];
	},

	/**
	 * @description Creates the list items that are shown in the result set.
	 * @returns {Array<JSX>} - Array of items created.
	 * @private
	 */
	_listItems() {
		if (!this.state.visible) return [];

		var list = [],
			set = this._getSet(),
			limit = set.length > FILE_LIMIT ? FILE_LIMIT : set.length;

		for (var i = 0; i < limit; i++) {
			var file = set[i],
				fileURL = router.fileURL(this.props.repo, this.props.rev, file);

			var ctx = classNames({
				selected: this.state.selectionIndex === i,
			});

			list.push(
				<li className={ctx} key={fileURL}>
					<a href={fileURL}>{file}</a>
				</li>
			);
		}

		return list;
	},

	render() {
		var ctx = classNames({
			"tree-entry-search": true,
			"hidden": !this.state.visible,
			"loading": this.state.loading,
		});

		return (
			<div className={ctx}>
				<div className="overlay" onClick={this._blurInput} />
				<div className="search-input-group">
					<div className="tree-search-input">
						<input type="text" onKeyUp={this._onType} placeholder="Search files in this repository..." />
						<div className="spinner"><i className="fa fa-spinner fa-spin" /></div>
					</div>
					<ul className="tree-search-file-list">
						{this._listItems()}
					</ul>
				</div>
			</div>
		);
	},
});

module.exports = TreeEntrySearch;
