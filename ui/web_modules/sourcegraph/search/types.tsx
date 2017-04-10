import { IPatternInfo } from "vs/platform/search/common/search";

export interface Query {
	query: SearchQuery;
}

export interface SearchQuery extends IPatternInfo {

}
