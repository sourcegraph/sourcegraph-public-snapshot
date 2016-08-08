import { IncomingMessage, ClientRequest } from 'http';
import Promise = require('any-promise');
import Request from './request';
import Response from './response';
export declare type Types = 'text' | 'buffer' | 'array' | 'uint8array' | 'stream' | string;
export interface Options {
    type?: Types;
    unzip?: boolean;
    jar?: any;
    agent?: any;
    maxRedirects?: number;
    rejectUnauthorized?: boolean;
    followRedirects?: boolean;
    confirmRedirect?: (request: ClientRequest, response: IncomingMessage) => boolean;
    ca?: string | Buffer | Array<string | Buffer>;
    cert?: string | Buffer;
    key?: string | Buffer;
    maxBufferSize?: number;
}
export declare function createTransport(options: Options): {
    use: ((request: Request, next: () => Promise<Response>) => Promise<Response>)[];
    abort: (request: Request) => void;
    open(request: Request): Promise<{}>;
};
