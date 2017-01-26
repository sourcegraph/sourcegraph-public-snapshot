import * as React from "react";
import { context } from "sourcegraph/app/context";
import { AuthDashboard } from "sourcegraph/dashboard";
import { NoAuthDashboard } from "sourcegraph/dashboard/NoAuthDashboard";
import { Home } from "sourcegraph/home/Home";

export function HomeRouter(props: any): JSX.Element {
	if (!context.authEnabled) {
		return <NoAuthDashboard {...props} />;
	} else if (context.user) {
		return <AuthDashboard {...props} />;
	}
	return <Home {...props} />;
}
