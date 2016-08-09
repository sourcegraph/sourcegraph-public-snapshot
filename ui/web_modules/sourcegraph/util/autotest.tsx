// tslint:disable

import * as TestUtils from "react-addons-test-utils";

import {mockTimeout} from "sourcegraph/util/mockTimeout";
import * as Dispatcher from "sourcegraph/Dispatcher";

import fs from "fs";
import child_process from "child_process";

let noJSON = undefined; // eslint-disable-line no-undefined

export function autotest(expected, filename, component, context?) {
	filename = `web_modules/${filename}`;
	// If fs is available, verify that expected matches the contents of filename.
	// If they don't match, it could be because the expected or filename arguemnts
	// are incorrectly specified. They should point to the same thing.
	if (fs.readFileSync) {
		let expectedOnDisk = fs.readFileSync(filename, {encoding: "utf-8"});
		if (JSON.stringify(expected, null, "\t") !== expectedOnDisk) {
			throw new Error(`autotest 'expected' argument doesn't match contents of ${filename} file, are you sure the first 2 arguments are correct?`);
		}
	}

	let renderer = TestUtils.createRenderer();
	let dispatchedToStores, dispatchedToBackends;
	dispatchedToStores = Dispatcher.Stores.catchDispatched(() => {
		dispatchedToBackends = Dispatcher.Backends.catchDispatched(() => {
			mockTimeout(() => {
				renderer.render(component, context);
			});
		});
	});
	if (dispatchedToStores.length !== 0) {
		throw new Error("do not dispatch to stores on render");
	}

	let json = JSON.stringify(
		{
			renderOutput: renderer.getRenderOutput(),
			dispatched: dispatchedToBackends.length > 0 ? dispatchedToBackends : noJSON,
		},
		(k, v) => {
			if ((k.charAt(0) === "_" && k !== "__html") || v === null || v === undefined) { // eslint-disable-line no-undefined
				return noJSON;
			}
			if (k === "children") {
				let children = toChildArray(v);
				if (children.length === 0) return noJSON;
				return mergeText(children);
			}
			switch (v.constructor) {
			case String:
			case Number:
			case Array:
			case Object:
			case Symbol: // eslint-disable-line no-undef
			case Boolean:
				return v;
			case Function:
				if (k === "type") {
					return v.name;
				}
				if (k.substr(0, 2) === "on") {
					let defaultPrevented: boolean | undefined = noJSON;
					let propagationStopped: boolean | undefined = noJSON;
					let funcDispatchedToStores, funcDispatchedToBackends;
					funcDispatchedToStores = Dispatcher.Stores.catchDispatched(() => {
						funcDispatchedToBackends = Dispatcher.Backends.catchDispatched(() => {
							mockTimeout(() => {
								v({
									preventDefault() {
										defaultPrevented = true;
									},
									stopPropagation() {
										propagationStopped = true;
									},
									currentTarget: {
										href: "[currentTarget.href]",
									},
									view: {
										scrollX: 11,
										scrollY: 22,
									},
									clientX: 10,
									clientY: 20,
								});
							});
						});
					});
					if (!defaultPrevented && funcDispatchedToStores.length === 0 && funcDispatchedToBackends.length === 0) {
						return noJSON;
					}
					return {
						defaultPrevented: defaultPrevented,
						propagationStopped: propagationStopped,
						dispatchedToStores: funcDispatchedToStores,
						dispatchedToBackends: funcDispatchedToBackends,
					};
				}
				return noJSON;
			default:
				return Object.assign({$constructor: v.constructor.name}, v);
			}
		},
		"\t"
	);

	if (JSON.stringify(expected, null, "\t") !== json) {
		if (fs.writeFileSync) {
			fs.writeFileSync(`${filename}.actual`, json);
			child_process.spawnSync("git", ["diff", "--no-index", filename, `${filename}.actual`], {stdio: [null, 1, 2]});
		}
		throw new Error("autotest mismatch");
	}
}

function toChildArray(children) {
	if (!children) {
		return [];
	}
	if (children.constructor !== Array) {
		return [removeDiv(children)];
	}
	return children
		.map(removeDiv)
		.map(toChildArray)
		.reduce((a, e) => a.concat(e), []);
}

function removeDiv(e) {
	if (e && e.type === "div" && Object.keys(e.props).length === 1 && e.props.children) {
		return e.props.children;
	}
	return e;
}

function mergeText(elements) {
	let merged: string[] = [];
	elements.forEach((e) => {
		if (e.constructor === Number) {
			e = String(e);
		}
		let i = merged.length - 1;
		if (e.constructor === String && i !== -1 && merged[i].constructor === String) {
			merged[i] += e;
			return;
		}
		merged.push(e);
	});
	return merged;
}
