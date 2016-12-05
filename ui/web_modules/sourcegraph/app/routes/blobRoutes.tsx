import { rel } from "sourcegraph/app/routePatterns";
import { Workbench } from "sourcegraph/workbench/workbench";

export const blobRoutes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			callback(null, { main: Workbench });
		},
		blobLoaderHelpers: [],
	},
];
