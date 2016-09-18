import * as React from "react";

import {colors} from "sourcegraph/components/jsStyles/colors";
import {input as inputStyle} from "sourcegraph/components/styles/input.css";
import {Search as SearchIcon} from "sourcegraph/components/symbols";

import {Category, SearchActions, categoryNames, deepLength} from "sourcegraph/search/modal/SearchContainer";
import {shortcuts} from "sourcegraph/search/modal/SearchModal";

const smallFont = 12.75;

const ResultRow = ({title, description, index, length, URLPath}, key, selected) => {
	let titleColor = colors.coolGray3();
	let backgroundColor = colors.coolGray1(.5);
	if (selected) {
		titleColor = colors.coolGray1();
		backgroundColor = colors.coolGray3();
	}

	return (
		<a key={key} style={{
			borderRadius: 3,
			padding: 16,
			margin: "0 8px 8px 8px",
			backgroundColor: backgroundColor,
			display: "block",
		}}
		onClick={() => console.log("actions.activateResult(URLPath)")}>
			{length ?
				<div>
				 <span style={{color: titleColor}}>{title.substr(0, index)}</span>
				 <span style={{color: colors.white(), fontWeight: "bold"}}>{title.substr(index, length)}</span>
				 <span style={{color: colors.coolGray3()}}>{title.substr(index + length)}</span>
				 </div> :
				 <div style={{color: colors.white()}}>
				 {title}
				 </div>
			}
			<div style={{fontSize: smallFont, color: colors.coolGray3()}}>
				{description}
			</div>
		</a>
	);
};

const ResultCategory = ({title, results, selected=-1}) => {
	if (results.length === 0) {
		return <div></div>;
	}

	const items = results.map((result, index) => {
		let isSelected = (index === selected);
		return ResultRow(result, index, isSelected);
	});

	return <div style={{padding: "14px 0"}}>
		<span style={{color: colors.coolGray3()}}>{title}</span>
		{
			results.map((result, index) => {
				return ResultRow(result, index, (index === selected));
			})
		}
	</div>;
}

export const ResultCategories = ({categories, limit, selection}) => {
	let sections = [];
	categories.forEach((category, i) => {
		let results = category.Results;
		let selected = -1;
		if (i === selection[0]) {
			selected = selection[1];
		}
		sections.push(<ResultCategory key={category.Title} title={category.Title} selected={selected} results={results} />);
	});
	return <div style={{overflow: "auto"}}>
		{sections}
	</div>;
};
