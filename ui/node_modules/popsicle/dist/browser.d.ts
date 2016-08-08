import Promise = require('any-promise');
import Request from './request';
import Response from './response';
export declare type Types = 'text' | 'document' | 'blob' | 'arraybuffer' | 'json' | string;
export interface Options {
    type?: Types;
    withCredentials?: boolean;
    overrideMimeType?: string;
}
export declare function createTransport(options: Options): {
    use: ((request: Request, next: () => Promise<Response>) => Promise<Response>)[];
    abort: (request: Request) => void;
    open(request: Request): Promise<Response>;
};
