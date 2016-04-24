declare module "amplitude-js" {
	declare function init(apiKey: string, user: ?Object, config: Object): void;
	declare function identify(op: any): void;
	declare function logEvent(eventName: string, props: Object): void;
}
