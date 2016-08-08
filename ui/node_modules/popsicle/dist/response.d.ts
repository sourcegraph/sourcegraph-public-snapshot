import Base, { BaseOptions, Headers, RawHeaders } from './base';
export interface ResponseOptions extends BaseOptions {
    body: any;
    status: number;
    statusText: string;
}
export interface ResponseJSON {
    headers: Headers;
    rawHeaders: RawHeaders;
    body: any;
    url: string;
    status: number;
    statusText: string;
}
export default class Response extends Base {
    status: number;
    statusText: string;
    body: any;
    constructor(options: ResponseOptions);
    statusType(): number;
    toJSON(): ResponseJSON;
}
