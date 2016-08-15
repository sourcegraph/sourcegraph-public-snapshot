import * as React from "react";
import {ResetPassword} from "sourcegraph/user/ResetPassword";
import {render} from "sourcegraph/util/renderTestUtils";

describe("ResetPassword", () => {
	it("should render initially", () => {
		render(<ResetPassword />, {signedIn: false});
	});
});
