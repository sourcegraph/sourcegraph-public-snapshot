import * as React from "react";
import { context } from "sourcegraph/app/context";
import { Dashboard } from "sourcegraph/dashboard";
import { Home } from "sourcegraph/home/Home";

export function HomeRouter(props: any): JSX.Element {
	return context.user ? <Dashboard {...props} /> : <Home {...props} />;
}
