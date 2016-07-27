import expect from "expect.js";
import {combineHeaders} from "sourcegraph/util/xhr";
import "whatwg-fetch";

function getHeaders(h: ?Headers): ?{[key: string]: string[]} {
	if (!h) return null;
	let m = {};
	const append = (name, val) => {
		if (!m[name]) m[name] = [];
		m[name].push(val);
	};

	if (h.forEach) {
		h.forEach((val, name) => append(name, val));
	} else {
		for (let [name, val] of h) {
			append(name, val);
		}
	}
	return m;
}

describe("combineHeaders", () => {
	it("should combine two Headers objects", () => {
		const a = new Headers({ // eslint-disable-line quote-props
			w: "w0",
			"x-y": "z0",
		});
		const b = new Headers({ // eslint-disable-line quote-props
			u: "u0",
			"X-Y": "z1",
		});

		expect(getHeaders(combineHeaders(a, b))).to.eql({ // eslint-disable-line quote-props
			w: ["w0"],
			u: ["u0"],
			"x-y": ["z0", "z1"],
		});
	});
});
