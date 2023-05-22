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
            debug: false,
            useContext: 'embeddings',
            experimentalSuggest: false,
            experimentalChatPredictions: false,
            experimentalGuardrails: false,
            experimentalInline: false,
            customHeaders: {},
            debugEnable: false,
            debugVerbose: false,
            debugFilter: '',
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
                    case 'cody.debug':
                        return true
                    case 'cody.useContext':
                        return 'keyword'
                    case 'cody.customHeaders':
                        return {
                            'Cache-Control': 'no-cache',
                            'Proxy-Authenticate': 'Basic',
                        }
                    case 'cody.experimental.suggestions':
                        return true
                    case 'cody.experimental.chatPredictions':
                        return true
                    case 'cody.experimental.guardrails':
                        return true
                    case 'cody.experimental.inline':
                        return true
                    case 'cody.experimental.enable':
                        return true
                    case 'cody.experimental.verbose':
                        return true
                    case 'cody.experimental.filter':
                        return ''
                    default:
                        throw new Error(`unexpected key: ${key}`)
                }
            },
        }
        expect(getConfiguration(config)).toEqual({
            serverEndpoint: 'http://example.com',
            codebase: 'my/codebase',
            debug: true,
            useContext: 'keyword',
            customHeaders: {
                'Cache-Control': 'no-cache',
                'Proxy-Authenticate': 'Basic',
            },
            experimentalSuggest: true,
            experimentalChatPredictions: true,
            experimentalGuardrails: true,
            experimentalInline: true,
            debugEnable: true,
            debugVerbose: true,
            debugFilter: '',
        })
    })
})
