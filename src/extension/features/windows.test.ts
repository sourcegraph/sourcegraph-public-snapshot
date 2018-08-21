import assert from 'assert'
import { MockMessageConnection } from '../../jsonrpc2/test/mockMessageConnection'
import {
    DidCloseTextDocumentNotification,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    ShowInputParams,
    ShowInputRequest,
} from '../../protocol'
import { Observable, Window, Windows } from '../api'
import { observableValue } from '../util'
import { createExtWindows } from './windows'

describe('ExtWindows', () => {
    function create(): { extWindows: Windows & Observable<Window[]>; mockConnection: MockMessageConnection } {
        const mockConnection = new MockMessageConnection()
        const extWindows = createExtWindows({ rawConnection: mockConnection })
        return { extWindows, mockConnection }
    }

    it('starts empty', () => {
        const { extWindows } = create()
        assert.deepStrictEqual(observableValue(extWindows), [{ isActive: true, activeComponent: null }] as Window[])
        assert.deepStrictEqual(extWindows.all, [{ isActive: true, activeComponent: null }] as Window[])
        assert.deepStrictEqual(extWindows.active, { isActive: true, activeComponent: null } as Window)
    })

    describe('component', () => {
        it('handles when a resource is opened', () => {
            const { extWindows, mockConnection } = create()
            mockConnection.recvNotification(DidOpenTextDocumentNotification.type, {
                textDocument: { uri: 'file:///a', languageId: 'l', text: 't', version: 1 },
            } as DidOpenTextDocumentParams)
            const expectedWindows: Window[] = [
                { isActive: true, activeComponent: { isActive: true, resource: 'file:///a' } },
            ]
            assert.deepStrictEqual(observableValue(extWindows), expectedWindows)
            assert.deepStrictEqual(extWindows.all, expectedWindows)
            assert.deepStrictEqual(extWindows.active, expectedWindows[0])
        })

        it('handles when the open resource is closed', () => {
            const { extWindows, mockConnection } = create()
            mockConnection.recvNotification(DidOpenTextDocumentNotification.type, {
                textDocument: { uri: 'file:///a', languageId: 'l', text: 't', version: 1 },
            } as DidOpenTextDocumentParams)
            mockConnection.recvNotification(DidCloseTextDocumentNotification.type, {
                textDocument: { uri: 'file:///a' },
            } as DidCloseTextDocumentParams)
            assert.deepStrictEqual(observableValue(extWindows), [{ isActive: true, activeComponent: null }] as Window[])
        })

        it('handles when a background resource is closed', () => {
            const { extWindows, mockConnection } = create()
            mockConnection.recvNotification(DidOpenTextDocumentNotification.type, {
                textDocument: { uri: 'file:///a', languageId: 'l', text: 't', version: 1 },
            } as DidOpenTextDocumentParams)
            mockConnection.recvNotification(DidCloseTextDocumentNotification.type, {
                textDocument: { uri: 'file:///b' },
            } as DidCloseTextDocumentParams)
            assert.deepStrictEqual(observableValue(extWindows), [
                { isActive: true, activeComponent: { isActive: true, resource: 'file:///a' } },
            ] as Window[])
        })
    })

    describe('showInputBox', () => {
        it('sends to the client', async () => {
            const { extWindows, mockConnection } = create()
            mockConnection.mockResults.set(ShowInputRequest.type.method, 'c')
            const input = await extWindows.showInputBox('a', 'b')
            assert.strictEqual(input, 'c')
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: ShowInputRequest.type.method,
                    params: { message: 'a', defaultValue: 'b' } as ShowInputParams,
                },
            ])
        })
    })
})
