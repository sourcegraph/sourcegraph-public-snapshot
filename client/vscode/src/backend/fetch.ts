import net from 'net'

import { ClientRequest, RequestOptions } from 'agent-base'
import HttpsProxyAgent from 'https-proxy-agent'
import fetch, { Headers } from 'node-fetch'
import vscode from 'vscode'

export { fetch, Headers }
export type { BodyInit, RequestInit, Response, HeadersInit } from 'node-fetch'

interface HttpsProxyAgentInterface {
    callback(req: ClientRequest, opts: RequestOptions): Promise<net.Socket>
}

type HttpsProxyAgentConstructor = new (
    options: string | { [key: string]: number | boolean | string | null }
) => HttpsProxyAgentInterface

export function getProxyAgent(): HttpsProxyAgentInterface | undefined {
    const proxyProtocol = vscode.workspace.getConfiguration('sourcegraph').get<string>('proxyProtocol')
    const proxyHost = vscode.workspace.getConfiguration('sourcegraph').get<string>('proxyHost')
    const proxyPort = vscode.workspace.getConfiguration('sourcegraph').get<number>('proxyPort')
    const proxyPath = vscode.workspace.getConfiguration('sourcegraph').get<string>('proxyPath')

    if (proxyProtocol || proxyHost || proxyPort || proxyPath) {
        return new ((HttpsProxyAgent as unknown) as HttpsProxyAgentConstructor)({
            ...(proxyProtocol ? { port: proxyProtocol } : null),
            ...(proxyHost ? { port: proxyHost } : null),
            ...(proxyPort ? { port: proxyPort } : null),
            ...(proxyPath ? { port: proxyPath } : null),
        })
    }

    return undefined
}
