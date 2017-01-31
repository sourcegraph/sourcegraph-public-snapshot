import expect from "expect.js";
import * as React from "react";
import { mockUser, mockUserAndGitHubToken } from "sourcegraph/app/context";
import { PricingPage } from "sourcegraph/page/PricingPage";
import { renderToString } from "sourcegraph/util/testutil/componentTestUtils";

describe("PricingPage", () => {
	it("should render for non-signed-in users", () => {
		mockUser(null, () => {
			let o = renderToString(<PricingPage location={{} as any} />);
			expect(o).to.not.contain("Full access during trial");
			expect(o).to.contain("Start 14 days free");
		});
	});
	it("should render for signed-in users without private code access", () => {
		mockUser({ UID: "1", Login: "Foo" }, () => {
			let o = renderToString(<PricingPage location={{} as any} />);
			expect(o).to.not.contain("Full access during trial");
			expect(o).to.contain("Start 14 days free");
		});
	});
	it("should render for signed-in users with private code access", () => {
		mockUserAndGitHubToken({ UID: "1", Login: "Foo" }, { scope: "read:org,repo,user:email" }, () => {
			let o = renderToString(<PricingPage location={{} as any} />);
			expect(o).to.contain("Full access during trial");
			expect(o).to.not.contain("Start 14 days free");
		});
	});

});
