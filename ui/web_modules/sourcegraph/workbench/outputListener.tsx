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

// In the main thread, we get a handle to the IOutputService (which
// deals with output channels) and can receive a callback whenever
// there is output on a given channel.
//
// This start function needs to be called after the IOutputService is
// set up. It is set up in sourcegraph/workbench/main's init function
// (by its call to setupServices, to be specific).
export function start(
	accessor: ServicesAccessor,
): void {
	const outputService = accessor.get<IOutputService>(IOutputService);

	// If you need to see when the output channel is created, and not
	// just when things are sent on it, you can use this to get the
	// signal:
	//
	// outputService.onOutputChannel(v => {
	// 	console.log("onOutputChannel", v);
	// });

	outputService.onOutput((event: IOutputEvent) => {
		const channel = outputService.getChannel(event.channelId!);
		OutputListener.dispatch(channel.label, event.output);
	});
};
