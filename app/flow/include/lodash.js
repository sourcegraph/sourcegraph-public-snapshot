declare module "lodash.debounce" {
	declare function debounce(func: (...args: any) => any, wait: ?number, options: {leading: boolean, trailing: boolean}): (...args: any) => any;
	declare var exports: typeof debounce;
}

declare module "lodash.trimleft" {
	declare function trimLeft(string: string, chars: ?string): string;
	declare var exports: typeof trimLeft;
}

declare module "lodash.get" {
	declare function get(object: Object, path: Array<string> | string, defaultValue: ?any): any;
	declare var exports: typeof get;
}
