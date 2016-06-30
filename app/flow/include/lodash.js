declare module "lodash.debounce" {
	declare function debounce(func: (...args: any) => any, wait: ?number, options: {leading: boolean, trailing: boolean}): (...args: any) => any;
	declare var exports: typeof debounce;
}

declare module "lodash.trimleft" {
	declare function trimLeft(string: string, chars: ?string): string;
	declare var exports: typeof trimLeft;
}

declare module "lodash/array/last" {
	declare function last(array: Array<any>): any;
	declare var exports: typeof last;
}

declare module "lodash/array/uniq" {
	declare function uniq(array: Array<any>): Array<any>;
	declare var exports: typeof uniq;
}

declare module "lodash/number/random" {
	declare function random(lower: number, upper: ?number, floating: ?boolean): any;
	declare var exports: typeof random;
}

declare module "lodash.get" {
	declare function get(object: Object, path: Array<string> | string, defaultValue: ?any): any;
	declare var exports: typeof get;
}
