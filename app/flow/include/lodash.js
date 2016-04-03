declare module "lodash/function/debounce" {
	declare function debounce(func: (...args: any) => any, wait: ?number, options: {leading: boolean, trailing: boolean}): (...args: any) => any;
	declare var exports: typeof debounce;
}

declare module "lodash/function/throttle" {
	declare function throttle(func: (...args: any) => any, wait: ?number, options?: {leading: boolean, trailing: boolean}): (...args: any) => any;
	declare var exports: typeof throttle;
}
