var React = require("react");

var Pagination = React.createClass({

	propTypes: {
		// currentPage is number corresponding to the page that is currently displayed.
		currentPage: React.PropTypes.number.isRequired,

		// onPageChange is the callback called when a new page is selected.
		onPageChange: React.PropTypes.func,

		// pageRange is the maximum number of page links that will be displayed in the
		// pagination menu.
		pageRange: React.PropTypes.number.isRequired,

		// totalPages is the total number of pages in the pagination list.
		totalPages: React.PropTypes.number.isRequired,
	},

	_onPageChange(page) {
		this.props.onPageChange(page);
	},

	// Determine the range of page numbers to display in the pagination menu.
	_calculatePageOffsets() {
		var firstPageOffset, lastPageOffset;
		// Using the current page as a midpoint, show half of the max page links to the
		// left and right of the current page.
		var leftPageCount = Math.ceil(this.props.pageRange/2);
		var rightPageCount = this.props.pageRange - leftPageCount;
		// Bound the first page offset to be at least 1.
		if (this.props.currentPage <= leftPageCount) {
			firstPageOffset = 1;
			lastPageOffset = Math.min(this.props.pageRange, this.props.totalPages);
		} else {
			// Calculate to offsets for the first and last page link that will be shown based on
			// the number of pages.
			firstPageOffset = this.props.currentPage - leftPageCount;
			firstPageOffset += 1; // Add 1 to account for showing the current page.
			// The offset for the last page is bounded by the total number of pages.
			lastPageOffset = Math.min(this.props.currentPage + rightPageCount, this.props.totalPages);
		}
		return [firstPageOffset, lastPageOffset];
	},

	render() {
		if (this.props.totalPages <= 1) return null;

		var pageOffsets = this._calculatePageOffsets();

		var pageList = [];
		for (var i=pageOffsets[0]; i<=pageOffsets[1]; i++) {
			pageList.push(
				<li key={i} className={i===this.props.currentPage ? "active" : ""}>
					<a onClick={this._onPageChange.bind(this, i)}>{i}</a>
				</li>
			);
		}

		return (
			<ul className="pagination">
				<li>
					<a aria-label="Previous"
						onClick={this._onPageChange.bind(this, 1)}>
						<span aria-hidden="true">&laquo;</span>
					</a>
				</li>
				{pageList}
				<li>
					<a aria-label="Previous"
						onClick={this._onPageChange.bind(this, this.props.totalPages)}>
						<span aria-hidden="true">&raquo;</span>
					</a>
				</li>
			</ul>
		);
	},
});

module.exports = Pagination;
