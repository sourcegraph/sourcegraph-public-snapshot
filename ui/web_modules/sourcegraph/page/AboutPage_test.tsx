import expect from "expect.js";
import * as React from "react";
import { mockUser } from "sourcegraph/app/context";
import { AboutPage } from "sourcegraph/page/AboutPage";
import { renderToString } from "sourcegraph/util/testutil/componentTestUtils";

describe("AboutPage", () => {
	it("should render for non-signed-in users", () => {
		mockUser(null, () => {
			let o = renderToString(<AboutPage location={{} as any} />);
			expect(o).to.contain("code intelligence");
			expect(o).to.contain("Quinn Slack");
		});
	});
	it("should render for signed-in users", () => {
		mockUser({ UID: "1", Login: "Foo" }, () => {
			let o = renderToString(<AboutPage location={{} as any} />);
			expect(o).to.contain("code intelligence");
			expect(o).to.contain("Quinn Slack");
		});
	});
});
