import * as sourcegraph from 'sourcegraph';
import { ClientCodeEditorAPI } from '../../client/api/codeEditor';
import { ClientWindowsAPI } from '../../client/api/windows';
import { ExtDocuments } from './documents';
export interface WindowData {
    visibleTextDocument: string | null;
}
/** @internal */
export interface ExtWindowsAPI {
    $acceptWindowData(allWindows: WindowData[]): void;
}
/** @internal */
export declare class ExtWindows implements ExtWindowsAPI {
    private windowsProxy;
    private codeEditorProxy;
    private documents;
    private data;
    /** @internal */
    constructor(windowsProxy: ClientWindowsAPI, codeEditorProxy: ClientCodeEditorAPI, documents: ExtDocuments);
    /** @internal */
    getActive(): sourcegraph.Window | undefined;
    /**
     * Returns all known windows.
     *
     * @internal
     */
    getAll(): sourcegraph.Window[];
    /** @internal */
    $acceptWindowData(allWindows: WindowData[]): void;
}
