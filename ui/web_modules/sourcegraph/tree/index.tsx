// tslint:disable: typedef ordered-imports

export function treeParam(splat: string[] | string): string {
	return splat instanceof Array ? splat[1] : "/";
}
