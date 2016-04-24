// Copied from latest master because it has more fetch-related APIs.
//
// Can remove this file when the release version contains Response, etc.

/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the "flow" directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 */
/* BOM */

// this part of spec is not finished yet, apparently
// https://stackoverflow.com/questions/35296664/can-fetch-get-object-as-headers
type HeadersInit = Headers | {[key: string]: string};


// TODO Heades and URLSearchParams are almost the same thing.
// Could it somehow be abstracted away?
declare class Headers {
    @@iterator(): Iterator<[string, string]>;
    constructor(init?: HeadersInit): void;
    append(name: string, value: string): void;
    delete(name: string): void;
    entries(): Iterator<[string, string]>;
    get(name: string): string;
    getAll(name: string): Array<string>;
    has(name: string): boolean;
    keys(): Iterator<string>;
    set(name: string, value: string): void;
    values(): Iterator<string>;
}

declare class URLSearchParams {
    @@iterator(): Iterator<[string, string]>;
    constructor(query?: string | URLSearchParams): void;
    append(name: string, value: string): void;
    delete(name: string): void;
    entries(): Iterator<[string, string]>;
    get(name: string): string;
    getAll(name: string): Array<string>;
    has(name: string): boolean;
    keys(): Iterator<string>;
    set(name: string, value: string): void;
    values(): Iterator<string>;
}

type CacheType =  'default' | 'no-store' | 'reload' | 'no-cache' | 'force-cache' | 'only-if-cached';
type CredentialsType = 'omit' | 'same-origin' | 'include';
type ModeType = 'cors' | 'no-cors' | 'same-origin';
type RedirectType = 'follow' | 'error' | 'manual';
type MethodType = 'GET' | 'POST' | 'DELETE' | 'HEAD' | 'OPTIONS' | 'PUT' ;
type ReferrerPolicyType =
    '' | 'no-referrer' | 'no-referrer-when-downgrade' | 'origin-only' |
    'origin-when-cross-origin' | 'unsafe-url';
type ResponseType =  'basic' | 'cors' | 'default' | 'error' | 'opaque' | 'opaqueredirect' ;

type RequestOptions = {
    body?: ?(Blob | FormData | URLSearchParams | string);

    cache?: ?CacheType;
    credentials?: ?CredentialsType;
	compress?: ?bool;
    headers?: ?HeadersInit;
    integrity?: ?string;
    method?: ?MethodType;
    mode?: ?ModeType;
    redirect?: ?RedirectType;
    referrer?: ?string;
    referrerPolicy?: ?ReferrerPolicyType;
}

type ResponseOptions = {
    status?: ?number;
    statusText?: ?string;
    headers?: ?HeadersInit
}

declare class Response {
    constructor(input?: string | URLSearchParams | FormData | Blob, init?: ResponseOptions): void;
    clone(): Response;
    static error(): Response;
    static redirect(url: string, status: number): Response;

    type: ResponseType;
    url: string;
    useFinalURL: boolean;
    ok: boolean;
    status: number;
    statusText: string;
    headers: Headers;

    // Body methods and attributes
    bodyUsed: boolean;

    arrayBuffer(): Promise<ArrayBuffer>;
    blob(): Promise<Blob>;
    formData(): Promise<FormData>;
    json(): Promise<any>;
    text(): Promise<string>;
}

declare class Request {
    constructor(input: string | Request, init?: RequestOptions): void;
    clone(): Request;

    url: string;

    cache: CacheType;
    credentials: CredentialsType;
    headers: Headers;
    integrity: string;
    method: MethodType;
    mode: ModeType;
    redirect: RedirectType;
    referrer: string;
    referrerPolicy: ReferrerPolicyType;

    // Body methods and attributes
    bodyUsed: boolean;

    arrayBuffer(): Promise<ArrayBuffer>;
    blob(): Promise<Blob>;
    formData(): Promise<FormData>;
    json(): Promise<any>;
    text(): Promise<string>;
}

declare function fetch(input: string | Request, init?: RequestOptions): Promise<Response>;
