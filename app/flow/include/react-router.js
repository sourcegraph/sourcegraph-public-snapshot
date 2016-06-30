import React from "react";

type Route = {
	path?: string;
	component?: React.Component | string;
	components?: {[key: string]: string | React.Component};
	getComponent?: (location: string, callback: (err: ?Error, component: React.Component) => void) => void;
	childRoutes?: Array<Route>;
	getChildRoutes?: (location: string, callback: (err: ?Error, routes: Array<Route>) => void) => void;
	indexRoute?: string;
	getIndexRoute?: (location: string, callback: (err: ?Error, route: Route) => void) => void;
};

type RouteParams = {[key: string]: string | string[]};

declare module "react-router" {
	declare var browserHistory: any;
	declare var match: Function;
	declare var RouterContext: React.Component;
	declare var Link: React.Component;

	declare type Route = Route;
	declare type RouteParams = RouteParams;

	declare type RouterState = {
		location: Location;
		routes: Array<Route>;
		params: RouteParams;
	};

	declare type RouterLocation = Location & {
		query: {[key: string]: string | string[]};
		state: {[key: string]: any};
	};
}

declare module "react-router/lib/PatternUtils" {
	declare function matchPattern(pattern: string, pathname: string): Object;
	declare function formatPattern(pattern: string, params: RouteParams): string;
	declare function getParams(pattern: string, params: RouteParams): {[key: string]: any};
}

declare module "react-router/lib/matchRoutes" {
	// routes arg type should be Array<Route> but that gives an unexpected error.
	declare function matchPattern(routes: Array<any>, location: {pathname: string}, cb: Function): void;
	declare var exports: typeof matchPattern;
}
