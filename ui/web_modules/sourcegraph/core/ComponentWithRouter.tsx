import * as React from "react";

import { RouterContext } from "sourcegraph/app/router";

export abstract class ComponentWithRouter<P, S> extends React.Component<P, S> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;
}
