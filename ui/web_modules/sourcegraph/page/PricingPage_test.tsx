import expect from "expect.js";
import * as React from "react";
import { mockUser } from "sourcegraph/app/context";
import { PricingPage } from "sourcegraph/page/PricingPage";
import { renderToString } from "sourcegraph/util/testutil/componentTestUtils";

describe("PricingPage", () => {
	it("should render for non-signed-in users", () => {
		mockUser(null, () => {
			let o = renderToString(<PricingPage />);
			expect(o).to.not.contain("Your current plan");
			expect(o).to.contain("Sign up");
		});
	});
	it("should render for signed-in users", () => {
		mockUser({ UID: "1", Login: "Foo" }, () => {
			let o = renderToString(<PricingPage />);
			expect(o).to.contain("Your current plan");
			expect(o).to.not.contain("Sign up");
		});
	});
});
