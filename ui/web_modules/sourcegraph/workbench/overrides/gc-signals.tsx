// No-op stub for the gc-signals npm package (which requires a native extension).

export class GCSignal {
	constructor(value: any) { /* noop */ }
}

export function consumeSignals(): number[] {
	return [];
}

export function onDidGarbageCollectSignals(callback: (ids: number[]) => any): { dispose(): void } {
	return { dispose(): void { /* noop */ } };
}

export function trackGarbageCollection(obj: any, id: number): number {
	return id;
}
