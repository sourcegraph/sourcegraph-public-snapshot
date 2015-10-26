import TestUtils from "react-addons-test-utils";

import mockTimeout from "./mockTimeout";
import Dispatcher from "../Dispatcher";

import fs from "fs";
import child_process from "child_process";

let noJSON = undefined; // eslint-disable-line no-undefined

export default function(expected, filename, component) {
	let renderer = TestUtils.createRenderer();
	let dispatched = Dispatcher.catchDispatched(() => {
		mockTimeout(() => {
			renderer.render(component);
		});
	});

	let json = JSON.stringify(
		{
			renderOutput: renderer.getRenderOutput(),
			dispatched: dispatched.length > 0 ? dispatched : noJSON,
		},
		(k, v) => {
			if (k.charAt(0) === "_" || v === null || v === undefined) { // eslint-disable-line no-undefined
				return noJSON;
			}
			if (k === "children") {
				return mergeText(toChildArray(v));
			}
			switch (v.constructor) {
			case String:
			case Number:
			case Array:
			case Object:
			case Symbol:
			case Boolean:
				return v;
			case Function:
				if (k === "type") {
					return v.name;
				}
				if (k.substr(0, 2) === "on") {
					let defaultPrevented = noJSON;
					let funcDispatched = Dispatcher.catchDispatched(() => {
						mockTimeout(() => {
							v({
								preventDefault() {
									defaultPrevented = true;
								},
							});
						});
					});
					if (!defaultPrevented && funcDispatched.length === 0) {
						return noJSON;
					}
					return {
						defaultPrevented: defaultPrevented,
						dispatched: funcDispatched,
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
		return [children];
	}
	return children.reduce((a, e) => a.concat(toChildArray(e)), []);
}

function mergeText(elements) {
	let merged = [];
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
