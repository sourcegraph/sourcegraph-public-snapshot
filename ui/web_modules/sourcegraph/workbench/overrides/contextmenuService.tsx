import { IContextMenuDelegate } from "vs/platform/contextview/browser/contextView";

export class ContextMenuService {
	// Disable
	public showContextMenu(delegate: IContextMenuDelegate): void { return; }
}
