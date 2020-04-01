declare module 'express-opentracing'

declare function middleware(options?: { tracer?: Tracer }): Handler
