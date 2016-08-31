import * as React from "react";
import {ForgotPassword} from "sourcegraph/user/ForgotPassword";
import {render} from "sourcegraph/util/testutil/renderTestUtils";

describe("ForgotPassword", () => {
	it("should render initially", () => {
		render(<ForgotPassword />, {signedIn: false});
	});
});
