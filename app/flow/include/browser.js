// https://github.com/facebook/flow/pull/1500/files adds this. When it's merged,
// we can remove this decl.
declare class Document_getSelection extends Document {
	getSelection(): ?Selection;
};
declare var document: Document_getSelection;
declare class Selection {
	rangeCount: number;
	getRangeAt(index: number): Range;
};

// Forward-ported from flow lib/dom.js commit 8bf0f53d432698525bc127bfbc5e642a4116a72f.
type EventHandler = (event: Event) => mixed
type EventListener = {handleEvent: EventHandler} | EventHandler
type KeyboardEventHandler = (event: KeyboardEvent) => mixed
type KeyboardEventListener = {handleEvent: KeyboardEventHandler} | KeyboardEventHandler

type KeyboardEventTypes = 'keydown' | 'keyup' | 'keypress';

declare class EventTarget {
    removeEventListener(type: KeyboardEventTypes, listener: KeyboardEventListener, useCapture?: boolean): void;
    addEventListener(type: KeyboardEventTypes, listener: KeyboardEventListener, useCapture?: boolean): void;
    detachEvent?: (type: KeyboardEventTypes, listener: KeyboardEventListener) => void;
    attachEvent?: (type: KeyboardEventTypes, listener: KeyboardEventListener) => void;

    removeEventListener(type: string, listener: EventListener, useCapture?: boolean): void;
    addEventListener(type: string, listener: EventListener, useCapture?: boolean): void;
    detachEvent?: (type: string, listener: EventListener) => void;
    attachEvent?: (type: string, listener: EventListener) => void;
    dispatchEvent(evt: Event): boolean;
}
