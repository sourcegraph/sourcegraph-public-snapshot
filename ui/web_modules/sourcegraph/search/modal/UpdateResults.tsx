import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {defaultFetch} from "sourcegraph/util/xhr";

import {Category, Result} from "sourcegraph/search/modal/SearchContainer";

import {TreeStore} from "sourcegraph/tree/TreeStore";

const fetch = singleflightFetch(defaultFetch);

// Update category finds the results for a category and calls the given callback
// with the results.
export function updateCategory(category: Category, query: string, callback: (results: Result[]) => void): void {
	if (category === Category.repository) {
		updateRepos(query, callback);
		return;
	} else if (category === Category.file) {
		callback(fileSearch(query));
	} else if (category === Category.definition) {
		const repository = "github.com/golang/go";
		updateDefs(repository, query, callback);
		return;
	}
}

function updateRepos(query: string, callback: (results: Result[]) => void): void {
	const url = `/.api/search-repos?Query=${query}`;
	fetch(url)
	.then(resp => resp.json())
	.then(data => data.Repos === undefined ? [] :
		data.Repos.map(({URI}) => ({title: URI})))
	.then(callback)
	.catch(error => console.error(error));
}

function updateDefs(repository: string, query: string, callback: (results: Result[]) => void): void {
	const url = `/.api/repos/${repository}/-/symbols?Query=${query}`;
	fetch(url)
	.then(resp => resp.json())
	.then(data => data.Repos === undefined ? [] :
		data.Repos.map(({URI}) => ({title: URI})))
	.then(callback)
	.catch(error => console.error(error));
}

function fileSearch(query: string): Result[] {
	let lowerQuery = query.toLowerCase();

	let results = new Array();
	if (lowerQuery === "") {
		return results;
	}
	// TODO: TreeStore.fileLists.get(this.props.repo, this.props.commitID);
	//       Need to pass the repo and commitID properties into this component, yeah?
	let fileList = TreeStore.fileLists.get("github.com/gorilla/mux", "0a192a193177452756c362c20087ddafcf6829c4");
	if (fileList === null) {
		return results;
	}
	for (let i = 0; i < fileList.Files.length; i++) {
		let index = fileList.Files[i].toLowerCase().indexOf(lowerQuery);
		if (index === -1) {
			continue;
		}
		results.push({title: fileList.Files[i], index: index, length: lowerQuery.length});
	}
	return results;
}
