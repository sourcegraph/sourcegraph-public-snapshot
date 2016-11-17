import {Action} from "redux";

export const SET_ACCESS_TOKEN = "SET_ACCESS_TOKEN";
export const RESOLVED_REV = "RESOLVED_REV";
export const FETCHED_ANNOTATIONS = "FETCHED_ANNOTATIONS";

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

export interface FetchedAnnotationsAction extends Action {
	repo: string;
	rev: string;
	path: string;
	relRev: string;
	xhrResponse: XHRResponse;
}

export interface AnnotationsResponse {
	Name: string; // path
	CommitID: string;
	Contents: string; // encoded
	IncludedAnnotations: {Annotations: Annotation[], LineStartBytes: number[]};
}

export interface Annotation {
	StartByte: number;
	EndByte: number;
	Class: string;
	WantInner: number;
}
