import React from "react";

class Pagination extends React.Component {
	constructor() {
		super();
		this._onPageChange = this._onPageChange.bind(this);
	}

	_onPageChange(page) {
		this.props.onPageChange(page);
	}

	// Determine the range of page numbers to display in the pagination menu.
	_calculatePageOffsets() {
		let firstPageOffset, lastPageOffset;
		// Using the current page as a midpoint, show half of the max page links to the
		// left and right of the current page.
		let rightPageCount = Math.floor((this.props.pageRange - 1)/2);
		let leftPageCount = this.props.pageRange - rightPageCount;
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
	}

	render() {
		if (this.props.totalPages <= 1) return null;

		let pageOffsets = this._calculatePageOffsets();

		let pageList = [];
		if (pageOffsets[0] > 1) {
			pageList.push(<li key="previous-indicator" className="disabled"><a>…</a></li>);
		}

		let i;
		for (i = pageOffsets[0]; i<=pageOffsets[1]; i++) {
			let pageLinkHTML;
			// If the current page is still loading, display a spnning indicator.
			if (i === this.props.currentPage && this.props.loading) {
				pageLinkHTML = <i className="fa fa-circle-o-notch fa-spin"></i>;
			} else {
				pageLinkHTML = i;
			}

			pageList.push(
				<li key={i} className={i===this.props.currentPage ? "active" : ""}>
					<a className="num-page-link"
						title={`Page ${i}`}
						onClick={this._onPageChange.bind(this, i)}>{pageLinkHTML}</a>
				</li>
			);
		}
		if (i < this.props.totalPages) {
			pageList.push(<li key="next-indicator" className="disabled"><a>…</a></li>);
		}

		let isFirstPage = this.props.currentPage === 1;
		let isLastPage = this.props.currentPage === this.props.totalPages;

		return (
			<ul className="pagination">
				<li key="first" className={isFirstPage ? "disabled" : null}>
					<a title={"Page 1"}
						onClick={isFirstPage ? null : this._onPageChange.bind(this, 1)}>
						<i className="fa fa-angle-double-left"></i>
					</a>
				</li>
				{pageList}
				<li key="last" className={isLastPage ? "disabled" : null}>
					<a title={`Page ${this.props.totalPages}`}
						onClick={isLastPage ? null : this._onPageChange.bind(this, this.props.totalPages)}>
						<i className="fa fa-angle-double-right"></i>
					</a>
				</li>
			</ul>
		);
	}
}

Pagination.propTypes = {
	currentPage: React.PropTypes.number.isRequired,
	onPageChange: React.PropTypes.func,
	// pageRange is the maximum number of page links that will be displayed in the
	// pagination menu.
	pageRange: React.PropTypes.number.isRequired,
	totalPages: React.PropTypes.number.isRequired,
	loading: React.PropTypes.bool,
};

export default Pagination;
