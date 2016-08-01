import Promise = require('any-promise');
import Request from '../request';
import Response from '../response';
export declare function wrap<T>(value: T): () => T;
export declare const headers: () => (request: Request, next: () => Promise<Response>) => Promise<Response>;
export declare const stringify: () => (request: Request, next: () => Promise<Response>) => Promise<Response>;
export declare type ParseType = 'json' | 'urlencoded';
export declare function parse(type: ParseType | ParseType[], strict?: boolean): (request: Request, next: () => Promise<Response>) => Promise<Response>;
