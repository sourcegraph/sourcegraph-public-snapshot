import { CancellationToken } from './cancel'
import { ResponseError } from './messages'

type HandlerResult<R, E> = R | ResponseError<E> | Promise<R> | Promise<ResponseError<E>> | Promise<R | ResponseError<E>>

export type StarRequestHandler = (method: string, ...params: any[]) => HandlerResult<any, any>

export type GenericRequestHandler<R, E> = (...params: any[]) => HandlerResult<R, E>

export type RequestHandler<P, R, E> = (params: P, token: CancellationToken) => HandlerResult<R, E>

export type StarNotificationHandler = (method: string, ...params: any[]) => void

export type GenericNotificationHandler = (...params: any[]) => void

export type NotificationHandler<P> = (params: P) => void
