import '@sourcegraph/testing/src/jestDomMatchers'
import fetch, { Headers, Request, Response } from 'node-fetch';

const global = globalThis as any;
global.fetch = fetch;
global.Headers = Headers;
global.Request = Request;
global.Response = Response;
