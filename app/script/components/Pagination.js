var React = require("react");

var Pagination = React.createClass({
	_onPageChange(page) {
		this.props.onPageChange(page);
	},

	render() {
		if (this.props.pages <= 1) return null;

		var prevPage = this.props.currentPage-1;
		var nextPage = this.props.currentPage+1;

		var pageList = [];
		for (var i=1; i<=this.props.pages; i++) {
			pageList.push(
				<li key={i} className={i===this.props.currentPage ? "active" : ""}>
					<a href={`#${i}`} onClick={this._onPageChange.bind(this, i)}>{i}</a>
				</li>
			);
		}

		return (
			<nav>
				<ul className="pagination">
					<li>
						<a href={`#${prevPage}`}
							aria-label="Previous"
							onClick={this._onPageChange.bind(this, prevPage)}>
							<span aria-hidden="true">&laquo;</span>
						</a>
					</li>
					{pageList}
					<li>
						<a href={`#${nextPage}`}
							aria-label="Previous"
							onClick={this._onPageChange.bind(this, nextPage)}>
							<span aria-hidden="true">&raquo;</span>
						</a>
					</li>
				</ul>
			</nav>
		);
	},
});

module.exports = Pagination;
