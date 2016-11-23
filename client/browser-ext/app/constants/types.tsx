import {Action} from "redux";

export const SET_ACCESS_TOKEN = "SET_ACCESS_TOKEN";
export const RESOLVED_REV = "RESOLVED_REV";

// TODO(john): consolidate this with default xhr Response type
export interface XHRResponse {
	status: number;
	body: any;
	json: Object;
}

export interface SetAccessTokenAction extends Action {
	token: string | null;
}

export interface ResolvedRevAction extends Action {
	repo: string;
	rev: string;
	xhrResponse: XHRResponse;
}
