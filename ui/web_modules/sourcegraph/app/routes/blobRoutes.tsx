import { rel } from "sourcegraph/app/routePatterns";
import { BlobMain } from "sourcegraph/blob/BlobMain";
import { RepoNavContext } from "sourcegraph/blob/RepoNavContext";
import { withLineColBoundToHash } from "sourcegraph/blob/withLineColBoundToHash";
import { withPath } from "sourcegraph/blob/withPath";
import { withResolvedRepoRev } from "sourcegraph/repo/withResolvedRepoRev";

let _blobMainComponent: any;

export const blobRoutes = [
	{
		path: rel.blob,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location: Location, callback: Function) => {
			if (!_blobMainComponent) {
				// Create only once to avoid unnecessary remounting after each route change.
				_blobMainComponent = withLineColBoundToHash(
					withResolvedRepoRev(
						withPath(
							BlobMain
						)
					)
				);
			}
			callback(null, {
				main: _blobMainComponent,
				repoNavContext: RepoNavContext,
			});
		},
		blobLoaderHelpers: [],
	},
];
