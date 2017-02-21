import { IOutputChannel } from "vs/workbench/parts/output/common/output";
import * as vs from "vscode/src/vs/workbench/parts/output/browser/outputServices";

export class OutputService extends vs.OutputService {
	public getChannel(id: string): IOutputChannel {
		const chan = super.getChannel(id);
		if (localStorage.getItem("logExtensionHostCommunication") === null) {
			// Disable the function that displays the output pane since it shouldn't be
			// user-facing.
			chan.show = (p?: boolean) => null as any;
		}
		return chan;
	}
}
