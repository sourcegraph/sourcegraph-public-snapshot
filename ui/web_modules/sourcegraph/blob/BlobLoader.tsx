// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Container} from "sourcegraph/Container";
import {BlobStore} from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";
import {DefStore} from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import "sourcegraph/tree/TreeBackend";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {withFileBlob} from "sourcegraph/blob/withFileBlob";
import {withAnnotations} from "sourcegraph/blob/withAnnotations";
import {BlobMain} from "sourcegraph/blob/BlobMain";
import * as invariant from "invariant";

export type Helper = {
	reconcileState: (state: any, props: any, context?: any) => void;
	onStateTransition?: (prevState: any, nextState: any) => void;
	renderProps?: (state: any) => any;
	render?: (state: any) => JSX.Element;
};

type Props = {
	route: any,
};

// blobLoader performs the portion of the work of loading a blob that differs based
// on what is originally being loaded. E.g., if we're loading a def, then the blob
// we show is the file in which the def is defined, so we must first fetch the def.
// If we're showing a file, then it's easier: we just show the blob of the file.
//
// BlobLoader is necessary to achieve good UI performance on the blob view.
//
// It is CRUCIAL for perf that the same React component tree be used for displaying
// the various kinds of blobs. We CAN'T have, for example, DefMain > Blob when
// viewing a def and BlobMain > Blob when showing a file. That's because React does
// not perform subtree matching (see https://facebook.github.io/react/docs/reconciliation.html#trade-offs).
// As a result, the Blob component in those 2 hierarchies would be unmounted (and all
// DOM nodes destroyed) if you click on defs in the UI and then go back to the file.
// This adds a noticeable lag and is unacceptable perf.
//
// That's why we use BlobLoader. It appears as the same component to React's
// reconciliation algorithm, but behind the scenes, it switches behavior (as described
// in the first paragraph) based on what route is loaded. It calls the methods of
// Helper (see the type defined above) at various stages of the component lifecycle.
// The helpers are obtained from the route definition, so they are different for
// files and defs.
//
// E.g., to see the helpers that get called by BlobLoader for a def, view the def
// routes file. The 3rd arg to the getComponents callback is defined (by us; it's not a
// standard react-router thing) to be the helpers used by the BlobLoader.
function blobLoader(Component) {
	class BlobLoader extends Container<Props, any> {
		_helpers: Array<Helper> | null;

		constructor(props: Props) {
			super(props);
			this._helpers = null;
		}

		reconcileState(state, props: Props) {
			if (props.route && state.route !== props.route) {
				// Clear state properties that were set by previous helpers.
				if (this._helpers) {
					this._helpers.forEach((h) => {
						if (h.reconcileState) {
							h.reconcileState(state, {}, this.context);
						}
					});
				}

				// This call is synchronous because we are guaranteed to already have
				// loaded these components' modules.
				invariant(props.route.blobLoaderHelpers, "route must define blobLoaderHelpers");
				this._helpers = props.route.blobLoaderHelpers;
			}

			Object.assign(state, props);

			if (this._helpers) {
				this._helpers.forEach((h) => {
					if (h.reconcileState) {
						h.reconcileState(state, props);
					}
				});
			}
		}

		onStateTransition(prevState, nextState) {
			if (this._helpers) {
				this._helpers.forEach((h) => {
					if (h.onStateTransition) {
						h.onStateTransition(prevState, nextState);
					}
				});
			}
		}

		// At least 1 store is required, so default to using BlobStore
		// since we've imported it anyway if we are here.
		//
		// TODO(sqs): dont require using all stores, take them from the helpers store fields
		stores() { return [DefStore, BlobStore, TreeStore]; }

		render(): JSX.Element | null {
			if (this._helpers) {
				for (let i = 0; i < this._helpers.length; i++) {
					const h = this._helpers[i];
					if (h.render) {
						const out = h.render(this.state);
						if (out) {
							return out;
						}
					}
				}
			}

			let renderProps = {};
			if (this._helpers) {
				this._helpers.forEach((h) => {
					if (h.renderProps) {
						Object.assign(renderProps, h.renderProps(this.state));
					}
				});
			}
			return <Component {...this.props} {...this.state} {...renderProps} />;
		}
	}

	return BlobLoader;
}

export const BlobLoader = (
	withResolvedRepoRev(
		blobLoader(
			withFileBlob(
				withAnnotations(
					BlobMain
				),
			),
		),
	)
);
