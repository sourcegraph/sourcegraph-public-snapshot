import * as React from "react";
import {ResetPassword} from "sourcegraph/user/ResetPassword";
import {render} from "sourcegraph/util/testutil/renderTestUtils";

describe("ResetPassword", () => {
	it("should render initially", () => {
		render(<ResetPassword />);
	});
});
