import * as React from "react";

import  {Heading} from "sourcegraph/components/Heading";
import {colors} from "sourcegraph/components/utils/colors";
import {modal_result, view_more} from "sourcegraph/quickopen/Style.css";

const smallFont = ".85rem";

const ViewMore = ({expandCategory, type}) => <a style={{
		textTransform: "uppercase",
		fontSize: smallFont,
		fontWeight: "bold",
		display: "block",
		textAlign: "center",
		marginTop: 16,
	}} className={view_more} onClick={expandCategory}>
	View more {type}
</a>;

export const ResultRow = ({title, description, index, length}, categoryIndex, itemIndex, selected, delegate, scrollIntoView) => {
	const oflow = {textOverflow: "ellipsis", overflow: "hidden"};
	return (
		<a key={itemIndex} className={modal_result}
			ref={node => { if (scrollIntoView && node && selected) {
			// Nonstandard, but improves experience in Chrome.
			if ((node as any).scrollIntoViewIfNeeded) {
				(node as any).scrollIntoViewIfNeeded(false);
			}
		}}}
		style={{
			borderRadius: 3,
			padding: "8px 20px",
			backgroundColor: selected ? colors.blue2() : colors.coolGray1(.5),
			display: "block",
			color:selected ? colors.white() : colors.coolGray4(),
			marginBottom: 8,
		}}
		onClick={() => delegate.select(categoryIndex, itemIndex)}>
		{length ? <div style={oflow}>
			<span>{title.substr(0, index)}</span>
			<span style={{fontWeight: "bold"}}>{title.substr(index, length)}</span>
			<span>{title.substr(index + length)}</span>
		</div> :
		<div style={oflow}>
			{title}
		</div>}
		<div style={Object.assign({fontSize: smallFont, color: selected ? colors.white(.7) : colors.coolGray3()}, oflow)}>
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
		<Heading color="gray"
		level={7}>
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
		return <div style={{padding: "14px 0", color: colors.white(), textAlign: "center"}}>No results found</div>;
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
	return <div style={{overflowY: "scroll"}}>
		{sections}
	</div>;
};

export const Hint = () => <div style={{color: colors.coolGray3(), margin: "8px auto", fontSize: smallFont}}>
	Press <span style={{
		backgroundColor: colors.coolGray1(.5),
		borderRadius: 3,
		padding: "2px 5px",
	}}>/</span> to open search from anywhere
</div>;
