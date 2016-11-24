import { rel } from "sourcegraph/app/routePatterns";
import { BlobMain } from "sourcegraph/blob/BlobMain";

export const blobRoutes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			callback(null, { main: BlobMain });
		},
		blobLoaderHelpers: [],
	},
];
