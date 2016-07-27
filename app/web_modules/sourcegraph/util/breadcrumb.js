export default function breadcrumb(path: string, sep: (key: number) => any, elemForPathComponent: (path: string, component: string, index: number, isLast: boolean) => any): Array<any> {
	let components = path.split("/");
	let elems = [];
	components.forEach((c, i) => {
		if (i !== 0) elems.push(sep(-1 * i));
		elems.push(elemForPathComponent(components.slice(0, i).join("/"), c, i, i === components.length - 1));
	});
	return elems;
}
