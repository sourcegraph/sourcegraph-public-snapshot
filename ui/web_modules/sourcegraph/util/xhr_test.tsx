import expect from "expect.js";
import "isomorphic-fetch";
import { combineHeaders } from "sourcegraph/util/xhr";

function getHeaders(h: Headers | null): { [key: string]: string[] } | null {
	if (!h) {
		return null;
	}
	let m = {};
	const append = (name, val) => {
		if (!m[name]) {
			m[name] = [];
		}
		m[name].push(val);
	};

	if (h.forEach) {
		h.forEach((val, name) => append(name, val));
	} else {
		for (let [name, val] of h as any) {
			append(name, val);
		}
	}
	return m;
}

describe("combineHeaders", () => {
	it("should combine two Headers objects", () => {
		const a = new Headers({
			w: "w0",
			"x-y": "z0",
		});
		const b = new Headers({
			u: "u0",
			"X-Y": "z1",
		});

		expect(getHeaders(combineHeaders(a, b))).to.eql({
			w: ["w0"],
			u: ["u0"],
			"x-y": ["z0", "z1"],
		});
	});
});
