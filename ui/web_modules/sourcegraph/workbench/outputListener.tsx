import { IDisposable } from "vs/base/common/lifecycle";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";
import { IOutputEvent, IOutputService } from "vs/workbench/parts/output/common/output";

// OutputListenerService relays output from the vscode OutputService to
// handlers that are supplied to it.
class OutputListenerService {
	private subs: Map<string, Set<(msg: string) => void>> = new Map();

	subscribe(name: string, handler: (msg: string) => void): IDisposable {
		const handlers = this.subs.get(name) || new Set();
		handlers.add(handler);
		this.subs.set(name, handlers);
		return {
			dispose: () => {
				const h = this.subs.get(name);
				if (h) { h.delete(handler); };
			}
		};
	}

	dispatch(name: string, msg: string): void {
		const handlers = this.subs.get(name);
		if (handlers) {
			handlers.forEach((f) => f(msg));
		}
	}
};

export const OutputListener = new OutputListenerService();

export function start(accessor: ServicesAccessor): void {
	const outputService = accessor.get<IOutputService>(IOutputService);

	outputService.onOutput((event: IOutputEvent) => {
		const channel = outputService.getChannel(event.channelId!);
		OutputListener.dispatch(channel.label, event.output);
	});
};
