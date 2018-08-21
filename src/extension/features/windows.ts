import { BehaviorSubject } from 'rxjs'
import { DidCloseTextDocumentNotification, DidOpenTextDocumentNotification, ShowInputRequest } from '../../protocol'
import { CXP, Observable, Window, Windows } from '../api'

class ExtWindows extends BehaviorSubject<Window[]> implements Windows, Observable<Window[]> {
    constructor(private ext: Pick<CXP<any>, 'rawConnection'>) {
        super([
            {
                isActive: true,
                activeComponent: null,
            },
        ])

        // Track last-opened text document.
        ext.rawConnection.onNotification(DidOpenTextDocumentNotification.type, params => {
            this.next([{ ...this.value[0], activeComponent: { isActive: true, resource: params.textDocument.uri } }])
        })
        ext.rawConnection.onNotification(DidCloseTextDocumentNotification.type, params => {
            if (
                this.active &&
                this.active.activeComponent &&
                this.active.activeComponent.resource &&
                this.active.activeComponent.resource === params.textDocument.uri
            ) {
                this.next([{ ...this.value[0], activeComponent: null }])
            }
        })
    }

    public get all(): Window[] {
        return this.value
    }

    public get active(): Window | null {
        return this.value.find(({ isActive }) => isActive) || null
    }

    public showInputBox(message: string, defaultValue?: string): Promise<string | null> {
        return this.ext.rawConnection.sendRequest(ShowInputRequest.type, { message, defaultValue })
    }

    public readonly [Symbol.observable] = () => this
}

/**
 * Creates the CXP extension API's {@link CXP#windows} value.
 *
 * @param ext The CXP extension API handle.
 * @return The {@link CXP#windows} value.
 */
export function createExtWindows<C>(ext: Pick<CXP<C>, 'rawConnection'>): Windows & Observable<Window[]> {
    return new ExtWindows(ext)
}
