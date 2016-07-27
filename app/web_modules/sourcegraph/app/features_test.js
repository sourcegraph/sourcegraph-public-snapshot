import * as React from "react";
import ReactDOMServer from "react-dom/server";
import expect from "expect.js";
import {setGlobalFeatures, withFeaturesContext} from "sourcegraph/app/features";

const C0base = (props, context) => <p>{context.features._testingDummyFeature}</p>; // eslint-disable-line react/jsx-key
C0base.contextTypes = {features: React.PropTypes.object.isRequired};
const C0 = withFeaturesContext(C0base);

describe("withFeaturesContext", () => {
	setGlobalFeatures({_testingDummyFeature: "bar"});

	it("should pass features in context to component", () => {
		let o = ReactDOMServer.renderToStaticMarkup(<C0/>);
		expect(o).to.eql("<p>bar</p>");
	});
});

