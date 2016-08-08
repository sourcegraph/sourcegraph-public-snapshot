import Promise = require('any-promise');
import Request from '../request';
import Response from '../response';
export * from './common';
export declare function headers(): (request: Request, next: () => Promise<Response>) => Promise<Response>;
