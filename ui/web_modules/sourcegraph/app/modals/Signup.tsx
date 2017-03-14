import * as React from "react";

import { SignupLoginAuth } from "sourcegraph/components";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { layout } from "sourcegraph/components/utils";

export function Signup(): JSX.Element {
	return <LocationStateModal title="Sign up" modalName="join" padded={false}>
		<SignupLoginAuth>
			To sign up, please authorize <br {...layout.hide.notSm } /> private code with GitHub:
			</SignupLoginAuth>
	</LocationStateModal>;
}
