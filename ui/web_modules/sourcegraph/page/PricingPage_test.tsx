import expect from "expect.js";
import * as React from "react";
import {PricingPage} from "sourcegraph/page/PricingPage";
import {renderToString} from "sourcegraph/util/componentTestUtils";

const dummyContext = {eventLogger: {logEvent: () => null}};

describe("PricingPage", () => {
	it("should render for non-signed-in users", () => {
		let o = renderToString(<PricingPage />, Object.assign({}, dummyContext, {signedIn: false}));
		expect(o).to.not.contain("Your current plan");
		expect(o).to.contain("Sign up");
	});
	it("should render for signed-in users", () => {
		let o = renderToString(<PricingPage />, Object.assign({}, dummyContext, {signedIn: true}));
		expect(o).to.contain("Your current plan");
		expect(o).to.not.contain("Sign up");
	});
});
