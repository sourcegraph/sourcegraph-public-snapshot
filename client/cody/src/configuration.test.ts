import type * as vscode from 'vscode'

import { getConfiguration } from './configuration'

describe('getConfiguration', () => {
    it('returns default values when no config set', () => {
        const config: Pick<vscode.WorkspaceConfiguration, 'get'> = {
            get: <T>(_key: string, defaultValue?: T): typeof defaultValue | undefined => defaultValue,
        }
        expect(getConfiguration(config)).toEqual({
            serverEndpoint: '',
            codebase: '',
            useContext: 'embeddings',
            completions: false,
            experimentalChatPredictions: false,
            experimentalGuardrails: false,
            experimentalInline: false,
            experimentalNonStop: false,
            customHeaders: {},
            debugEnable: false,
            debugVerbose: false,
            debugFilter: null,
        })
    })

    it('reads values from config', () => {
        const config: Pick<vscode.WorkspaceConfiguration, 'get'> = {
            get: key => {
                switch (key) {
                    case 'cody.serverEndpoint':
                        return 'http://example.com'
                    case 'cody.codebase':
                        return 'my/codebase'
                    case 'cody.useContext':
                        return 'keyword'
                    case 'cody.customHeaders':
                        return {
                            'Cache-Control': 'no-cache',
                            'Proxy-Authenticate': 'Basic',
                        }
                    case 'cody.completions':
                    case 'cody.experimental.suggestions':
                        return true
                    case 'cody.experimental.chatPredictions':
                        return true
                    case 'cody.experimental.guardrails':
                        return true
                    case 'cody.experimental.inline':
                        return true
                    case 'cody.experimental.nonStop':
                        return true
                    case 'cody.debug.enable':
                        return true
                    case 'cody.debug.verbose':
                        return true
                    case 'cody.debug.filter':
                        return /.*/
                    default:
                        throw new Error(`unexpected key: ${key}`)
                }
            },
        }
        expect(getConfiguration(config)).toEqual({
            serverEndpoint: 'http://example.com',
            codebase: 'my/codebase',
            useContext: 'keyword',
            customHeaders: {
                'Cache-Control': 'no-cache',
                'Proxy-Authenticate': 'Basic',
            },
            completions: true,
            experimentalChatPredictions: true,
            experimentalGuardrails: true,
            experimentalInline: true,
            experimentalNonStop: true,
            debugEnable: true,
            debugVerbose: true,
            debugFilter: /.*/,
        })
    })
})
