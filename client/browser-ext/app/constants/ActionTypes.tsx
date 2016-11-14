import {Action} from "redux";

export const SET_ACCESS_TOKEN = "SET_ACCESS_TOKEN";
export const RESOLVED_REV = "RESOLVED_REV";
export const FETCHED_ANNOTATIONS = "FETCHED_ANNOTATIONS";

export interface XHRResponse {
	status: number;
	body: any; // TODO(john): consolidate types
	json: Object;
}

// TODO(john): make these classes
export interface SetAccessTokenAction extends Action {
	token: string | null;
}

export interface ResolvedRevAction extends Action {
	repo: string;
	rev: string;
	xhrResponse: XHRResponse;
}

export interface FetchedAnnotationsAction extends Action {
	repo: string;
	rev: string;
	path: string;
	relRev: string;
	xhrResponse: XHRResponse;
}
