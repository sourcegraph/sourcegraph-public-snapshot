import URI from "vs/base/common/uri";
import { IEditorInput } from "vs/platform/editor/common/editor";

export function getResource(input: IEditorInput): URI {
	if (input["resource"]) {
		return (input as any).resource;
	} else {
		throw "Couldn't find resource.";
	}
}

export const NoopDisposer = { dispose: () => {/* */ } };
