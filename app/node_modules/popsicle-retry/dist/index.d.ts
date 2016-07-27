import Promise = require('any-promise');
import { Request, Response, PopsicleError } from 'popsicle';
declare function popsicleRetry(retries?: (error: PopsicleError, response: Response, iter: number) => number): (request: Request, next: () => Promise<Response>) => Promise<any>;
declare namespace popsicleRetry {
    function retryAllowed(error: PopsicleError, response: Response): boolean;
    function retries(count?: number, isRetryAllowed?: typeof retryAllowed): (error: PopsicleError, response: Response, iter: number) => number;
}
export = popsicleRetry;
