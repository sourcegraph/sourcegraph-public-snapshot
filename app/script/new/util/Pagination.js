import React from "react";

import Component from "../Component";

class Pagination extends Component {
	constructor(props) {
		super(props);
		this._onPageChange = this._onPageChange.bind(this);
	}

	_onPageChange(page) {
		this.state.onPageChange(page);
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		// Determine the range of page numbers to display in the pagination menu.
		//
		// Using the current page as a midpoint, show half of the max page links to the
		// left and right of the current page.
		let rightPageCount = Math.floor((props.pageRange - 1)/2);
		let leftPageCount = props.pageRange - rightPageCount;
		// Bound the first page offset to be at least 1.
		if (props.currentPage <= leftPageCount) {
			state.firstPageOffset = 1;
			state.lastPageOffset = Math.min(props.pageRange, props.totalPages);
		} else {
			// Calculate to offsets for the first and last page link that will be shown based on
			// the number of pages.
			state.firstPageOffset = props.currentPage - leftPageCount;
			state.firstPageOffset += 1; // Add 1 to account for showing the current page.
			// The offset for the last page is bounded by the total number of pages.
			state.lastPageOffset = Math.min(props.currentPage + rightPageCount, props.totalPages);
		}
	}

	render() {
		if (this.state.totalPages <= 1) return null;

		let pageList = [];
		if (this.state.firstPageOffset > 1) {
			pageList.push(<li key="previous-indicator" className="disabled"><a>…</a></li>);
		}

		let i;
		for (i = this.state.firstPageOffset; i<=this.state.lastPageOffset; i++) {
			let pageLinkHTML;
			// If the current page is still loading, display a spnning indicator.
			if (i === this.state.currentPage && this.state.loading) {
				pageLinkHTML = <i className="fa fa-circle-o-notch fa-spin"></i>;
			} else {
				pageLinkHTML = i;
			}

			pageList.push(
				<li key={i} className={i===this.state.currentPage ? "active" : ""}>
					<a className="num-page-link"
						title={`Page ${i}`}
						onClick={this._onPageChange.bind(this, i)}>{pageLinkHTML}</a>
				</li>
			);
		}
		if (i < this.state.totalPages) {
			pageList.push(<li key="next-indicator" className="disabled"><a>…</a></li>);
		}

		let isFirstPage = this.state.currentPage === 1;
		let isLastPage = this.state.currentPage === this.state.totalPages;

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
					<a title={`Page ${this.state.totalPages}`}
						onClick={isLastPage ? null : this._onPageChange.bind(this, this.state.totalPages)}>
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
