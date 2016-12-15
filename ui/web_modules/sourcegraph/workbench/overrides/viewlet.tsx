import * as orig from "vscode/src/vs/workbench/browser/viewlet";

import { VIEWLET_ID} from "vs/workbench/parts/files/common/files";

export class ViewletRegistry extends orig.ViewletRegistry {
	getDefaultViewletId(): string {
		return VIEWLET_ID;
	}
}

export const Viewlet = orig.Viewlet;
export const ViewerViewlet = orig.ViewerViewlet;
export const ViewletDescriptor = orig.ViewletDescriptor;
export const Extensions = orig.Extensions;
export const ToggleViewletAction = orig.ToggleViewletAction;
export const CollapseAction = orig.CollapseAction;
export const AdaptiveCollapsibleViewletView = orig.AdaptiveCollapsibleViewletView;
export const CollapsibleViewletView = orig.CollapsibleViewletView;
