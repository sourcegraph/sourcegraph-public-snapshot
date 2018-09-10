import assert from 'assert'
import { MockMessageConnection } from '../../../jsonrpc2/test/mockMessageConnection'
import {
    DidCloseTextDocumentNotification,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    ShowInputParams,
    ShowInputRequest,
} from '../../../protocol'
import { URI } from '../../../types/uri'
import { Window } from '../api'
import { observableValue } from '../util'
import { ExtWindows } from './windows'

describe('ExtWindows', () => {
    function create(): { extWindows: ExtWindows; mockConnection: MockMessageConnection } {
        const mockConnection = new MockMessageConnection()
        const extWindows = new ExtWindows(mockConnection)
        return { extWindows, mockConnection }
    }

    it('starts empty', () => {
        const { extWindows } = create()
        assert.deepStrictEqual(observableValue(extWindows), [{ isActive: true, activeComponent: null }] as Window[])
        assert.deepStrictEqual(extWindows.activeWindow, { isActive: true, activeComponent: null } as Window)
    })

    describe('component', () => {
        it('handles when a resource is opened', () => {
            const { extWindows, mockConnection } = create()
            mockConnection.recvNotification(DidOpenTextDocumentNotification.type, {
                textDocument: { uri: 'file:///a', languageId: 'l', text: 't', version: 1 },
            } as DidOpenTextDocumentParams)
            const expectedWindows: Window[] = [
                { isActive: true, activeComponent: { isActive: true, resource: URI.parse('file:///a') } },
            ]
            assert.deepStrictEqual(observableValue(extWindows), expectedWindows)
            assert.deepStrictEqual(extWindows.activeWindow, expectedWindows[0])
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
                { isActive: true, activeComponent: { isActive: true, resource: URI.parse('file:///a') } },
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
