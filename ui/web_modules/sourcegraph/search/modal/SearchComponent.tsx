import * as React from "react";

import {Input} from "sourcegraph/components/Input";
import {colors} from "sourcegraph/components/jsStyles/colors";

import {Search as SearchIcon} from "sourcegraph/components/symbols";

import {Category, Result, SearchActions, categoryNames, deepLength} from "sourcegraph/search/modal/SearchContainer";
import {shortcuts} from "sourcegraph/search/modal/SearchModal";

const smallFont = 12.75;

const modalStyle = {
	position: "fixed",
	top: 60,
	right: 0,
	left: 0,
	maxWidth: 515,
	margin: "0 auto",
	borderRadius: "0 0 8px 8px",
	backgroundColor: colors.coolGray2(),
	padding: 16,
	display: "flex",
	flexDirection: "column",
	zIndex: 1,
	maxHeight: "90vh",
	fontSize: 15,
};

const Bubble = (props) => <span style={{
	backgroundColor: colors.coolGray1(),
	borderRadius: 3,
	padding: "2px 5px",
	textTransform: "uppercase"}}>
	<b>{props.children}</b>
</span>;

const Hint = ({tag}) => {
	let keycode;
	switch (tag) {
		case Category.definition:
			keycode = <Bubble>{shortcuts.def}</Bubble>;
			break;
		case Category.repository:
			keycode = <Bubble>{shortcuts.repo}</Bubble>;
			break;
		case Category.file:
			keycode = <Bubble>{shortcuts.file}</Bubble>;
			break;
		case null:
			keycode = <Bubble>⌘ shift p</Bubble>;
			break;
	}
	return <div style={{color: colors.coolGray3(), margin: "8px auto", fontSize: smallFont}}>
		Hit {keycode} to bring this up from anywhere
	</div>;
};

const Result = ({title, description, index, length}, key) => <div key={key} style={{
	backgroundColor: colors.coolGray1(.5),
	borderRadius: 3,
	padding: 16,
	margin: "0 8px 8px 8px",
}}>
	{length ? <div>
			<span style={{color: colors.coolGray3()}}>{title.substr(0, index)}</span>
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
</div>;

const ViewAll = ({noun, viewCategory}) => <div style={{
	color: colors.blueText(),
	textAlign: "center",
	fontSize: smallFont,
	textTransform: "uppercase",
}}>
	<b><a onClick={viewCategory}>view all {noun}</a></b>
</div>;

const ResultCategory = ({category, results, actions, limit = Infinity}) => {
	if (results.length === 0) {
		return <div></div>;
	}
	let title;
	let noun;
	const plural = results.length > 1;
	switch (category) {
		case Category.definition:
			noun = plural ? "definitions" : "definition";
			title = `${results.length} ${noun} in this repository`;
			break;
		case Category.repository:
			noun = plural ? "repositories" : "repositories";
			title = `${results.length} ${noun}`;
			break;
		case Category.file:
			noun = plural ? "files" : "file";
			title = `${results.length} ${noun} in this repository`;
			break;
	}
	let items;
	if (results.length > limit) {
		results = results.slice(0, limit);
		items = results.map(Result);
		items.push(<ViewAll key={limit} noun={noun} viewCategory={() => actions.viewCategory(category)} />);
	} else {
		items = results.map(Result);
	}
	return <div key={category} style={{padding: "14px 0"}}>
		<span style={{color: colors.coolGray3()}}>{title}</span>
		{items}
	</div>;
};

const ResultCategories = ({resultCategories, limit, actions}) => {
	let categoryLimit = Infinity;
	if (deepLength(resultCategories) > limit) {
		categoryLimit = 5;
	}
	let sections = new Array();
	resultCategories.forEach((results, category) => {
		sections.push(<ResultCategory key={category} category={category} actions={actions} limit={categoryLimit} results={results} />);
	});
	return <div style={{overflow: "scroll"}}>
		{sections}
	</div>;
};

const Tag = ({tag}) => {
	if (tag === null) {
		return <div></div>;
	}
	let content;
	switch (tag) {
			case Category.definition:
			content = "def";
		break;
			case Category.repository:
			content = "repo";
		break;
		case Category.file:
			content = "file";
			break;
	}
	return <div style={{backgroundColor: colors.coolGray4(), margin: "0 5px", padding: "0 5px", borderRadius: 3}}>
		{content}:
	</div>;
};

const SearchInput = ({tag, input, actions}) => <div style={{
	backgroundColor: colors.white(),
	borderRadius: 3,
	width: "100%",
	padding: "3px 10px",
	flex: "0 0 auto",
	height: 45,
	display: "flex",
	alignItems: "center",
	flexDirection: "row",
}}>
	<SearchIcon style={{fill: colors.coolGray2()}}/>
	<Tag tag={tag} />
	<Input
		style={{boxSizing: "border-box", border: "none", flex: "1 0 auto"}}
		placeholder="new http request"
		value={input}
		ref={actions.bindSearchInput}
		block={true}
		onChange={actions.updateInput} />
	<button onClick={actions.dismiss} style={{display: "inline"}}>x</button>
</div>;

const CategorySelection = ({title, content, shortcut, selected}: {title: string, content: string, shortcut: string, selected: boolean}) => <div style={{
	padding: 5,
	color: colors.coolGray3(),
	backgroundColor: selected ? colors.blue2() : "transparent",
	borderRadius: 3,
}}>
	<b style={{color: colors.white(), marginRight: 8}}>{title}:</b>
	<span style={{color: selected ? colors.white() : colors.coolGray3()}}>{content}</span>
	<span style={{color: colors.white(), float: "right"}}>
		<Bubble>{shortcut}</Bubble>
	</span>
</div>;

const CategorySelector = ({sel}: {sel: number}) => <div>
	<span style={{color: colors.coolGray3(), fontSize: smallFont}}>JUMP TO ...</span>
	<CategorySelection title={"repo"} content={"repository name"} shortcut={shortcuts.repo} selected={sel === 1}/>
	<CategorySelection title={"file"} content={"filename in this repo"} shortcut={shortcuts.file} selected={sel === 2} />
	<CategorySelection title={"def"} content={"definition in this repo"} shortcut={shortcuts.def} selected={sel === 3} />
</div>;

const Tab = ({category, count, selected, actions}: {category: Category, count: number, selected: boolean, actions: SearchActions}) => {
	const catNames = categoryNames.get(category);
	if (!catNames) { throw new Error("category names not set"); }
	const catName = count > 1 ? catNames[1] : catNames[0];
	if (selected) {
		return <span style={{
			color: colors.blue3(),
			padding: 8,
			display: "inline-block",
			borderBottomStyle: "solid",
			borderBottomWidth: 3,
		}}>
			{count} {catName}
		</span>;
	} else {
		return <a style={{
			color: colors.coolGray3(),
			padding: 8,
			display: "inline-block",
		}}
		onClick={() => actions.viewCategory(category)}>
			{count} {catName}
		</a>;
	}
};

const Tabs = ({actions, tab, categories}) => {
	let tabs = new Array();
	categories.forEach((results, category) => {
		let selected = false;
		if (category === tab) {
			selected = true;
		}
		if (results.length > 0) {
			tabs.push(<Tab key={category} category={category} count={results.length} selected={selected} actions={actions}/>);
		}
	});
	return <div style={{textAlign: "center"}}>
		{tabs}
	</div>;
};

const ResultList = ({results}) => <div style={{
	overflow: "scroll",
}}>
	{results.map(Result)}
</div>;

const TabbedResults = ({actions, tab, results}) => <div style={{
	display: "flex",
	flexDirection: "column",
}}>
	<Tabs actions={actions} tab={tab} categories={results}/>
	<ResultList results={results.get(tab)} />
</div>;

interface ComponentData {
	tag: Category | null;
	tab: Category | null;
	input: string;
	selected: number;
	results: Map<Category, Result[]>;
}

interface Props {
	data: ComponentData;
	actions: SearchActions;
}

const SingleCategoryResults = ({data, actions, category}) => {
	const results = data.results.get(category);
	return <div style={{overflow: "scroll"}}>
		<ResultCategory category={category} actions={actions} results={results} />
	</div>;
};

// SearchComponent contains the the view logic.
export const SearchComponent = ({data, actions}: Props) => {
	let content;
	let showHint = true;
	if (data.input === "" && data.tag === null) {
		content = <CategorySelector sel={data.selected} />;
	} else if (data.tag !== null) {
		content = <SingleCategoryResults actions={actions} data={data} category={data.tag} />;
	} else if (data.tab !== null) {
		content = <TabbedResults actions={actions} tab={data.tab} results={data.results}/>;
		showHint = false;
	} else {
		content = <ResultCategories resultCategories={data.results} limit={15} actions={actions}/>;
	}
	return <div style={modalStyle}>
		<SearchInput tag={data.tag} input={data.input} actions={actions} />
		{showHint && <Hint tag={data.tag} />}
		{content}
	</div>;
};
