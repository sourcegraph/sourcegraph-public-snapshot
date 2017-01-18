import { hover } from "glamor";
import * as React from "react";

import { Heading } from "sourcegraph/components/Heading";
import { colors, typography } from "sourcegraph/components/utils";

const ViewMore = ({expandCategory, type}) => <a style={Object.assign({
	color: colors.blueL1(),
	textTransform: "uppercase",
	fontWeight: "bold",
	display: "block",
	textAlign: "center",
	marginTop: 16,
}, typography.small)} {...hover({ color: `${colors.blueL2()} !important` }) } onClick={expandCategory}>
	View more {type}
</a>;

export const ResultRow = ({title, description, index, length}, categoryIndex, itemIndex, selected, delegate, scrollIntoView) => {
	const oflow = { textOverflow: "ellipsis", overflow: "hidden" };
	return (
		<a
			key={itemIndex}
			data-class-name={selected ? "modal-result-selected modal-result" : "modal-result"}
			ref={node => {
				if (scrollIntoView && node && selected) {
					// Nonstandard, but improves experience in Chrome.
					if ((node as any).scrollIntoViewIfNeeded) {
						(node as any).scrollIntoViewIfNeeded(false);
					}
				}
			} }
			style={{
				borderRadius: 3,
				padding: "8px 20px",
				border: "2px solid transparent",
				backgroundColor: selected ? colors.blue() : colors.blueGrayD2(0.5),
				display: "block",
				color: selected ? colors.white() : colors.blueGrayL3(),
				marginBottom: 8,
			}}
			{...hover({ border: `2px solid ${colors.blue()} !important` }) }
			onClick={() => delegate.select(categoryIndex, itemIndex)}>
			{length ? <div style={oflow}>
				<span>{title.substr(0, index)}</span>
				<span style={{ fontWeight: "bold" }}>{title.substr(index, length)}</span>
				<span>{title.substr(index + length)}</span>
			</div> :
				<div style={oflow}>
					{title}
				</div>}
			<div style={Object.assign({ color: selected ? colors.white(0.7) : colors.blueGrayL1() }, oflow, typography.small)}>
				{description}
			</div>
		</a>
	);
};

const ResultCategory = ({title, results, selected, delegate, categoryIndex, limit, expandCategory, scrollIntoView}) => {
	const total = results.length;
	if (total === 0) {
		return <div></div>;
	}
	results = results.slice(0, limit);
	return <div>
		<Heading level={7} style={{ color: colors.blueGrayL1() }}>
			{title}
		</Heading>
		{results.map((result, index) =>
			ResultRow(result, categoryIndex, index, (index === selected), delegate, scrollIntoView)
		)}
		{total > limit ? <ViewMore expandCategory={expandCategory} type={title} /> : null}
	</div>;
};

export const ResultCategories = ({categories, selection, delegate, limits, scrollIntoView}) => {
	const loadingOrFound = categories.some(cat => cat.IsLoading || cat.Results.length > 0);
	if (!loadingOrFound) {
		return <div style={{ padding: "14px 0", color: colors.white(), textAlign: "center" }}>No results found</div>;
	}
	let sections: JSX.Element[] = [];
	categories.forEach((category, i) => {
		let selected = -1;
		if (i === selection.category) {
			selected = selection.row;
		}
		sections.push(<ResultCategory
			key={category.Title}
			limit={limits[i]}
			categoryIndex={i}
			title={category.Title}
			selected={selected}
			results={category.Results}
			delegate={delegate}
			scrollIntoView={scrollIntoView}
			expandCategory={delegate.expand(i)} />);
	});
	return <div style={{ overflowY: "scroll" }}>
		{sections}
	</div>;
};

export const Hint = () => <div style={Object.assign({ color: colors.blueGrayL1(), margin: "8px auto" }, typography.small)}>
	Press <span style={{
		backgroundColor: colors.blueGrayD2(.5),
		borderRadius: 3,
		padding: "2px 5px",
	}}>/</span> to open search from anywhere
</div>;
