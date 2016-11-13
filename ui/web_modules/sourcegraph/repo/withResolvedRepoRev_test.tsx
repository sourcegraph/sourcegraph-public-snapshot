import * as React from "react";
import {Location} from "sourcegraph/Location";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {render} from "sourcegraph/util/testutil/renderTestUtils";

const C = withResolvedRepoRev((props) => null);

describe("withResolvedRepoRev", () => {
	it("should render initially", () => {
		render(<C params={{splat: "r"}} location={{} as Location} />);
	});
});
