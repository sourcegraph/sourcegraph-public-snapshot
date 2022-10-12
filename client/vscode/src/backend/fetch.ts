import HttpsProxyAgent from 'https-proxy-agent'
import fetch, { Headers } from 'node-fetch'

export { fetch, Headers }
export type { BodyInit, RequestInit, Response, HeadersInit } from 'node-fetch'

export function getProxyAgent(): any | undefined {
    // const proxyUrl = 'http://localhost:1337'

    // if (proxyUrl) {
    // @ts-ignore
    const agent = new HttpsProxyAgent({
        protocol: 'http',
        // host: '192.168.0.220',
        // port: '9090',
        path: '/Users/philipp/dev/domain-socket-proxy/unix.socket',
    })
    return agent
    // }

    // if (strictSSL === false) {
    //     return new HttpProxyAgent({
    //         rejectUnauthorized: false,
    //     })
    // }

    // return undefined
}
