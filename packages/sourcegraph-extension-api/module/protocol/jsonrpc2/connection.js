import { CancelNotification } from './cancel';
import { CancellationToken, CancellationTokenSource } from './cancel';
import { Emitter } from './events';
import { LinkedMap } from './linkedMap';
import { ErrorCodes, isNotificationMessage, isRequestMessage, isResponseMessage, ResponseError, } from './messages';
import { noopTracer, Trace } from './trace';
const NullLogger = Object.freeze({
    error: () => {
        /* noop */
    },
    warn: () => {
        /* noop */
    },
    info: () => {
        /* noop */
    },
    log: () => {
        /* noop */
    },
});
export var ConnectionErrors;
(function (ConnectionErrors) {
    /**
     * The connection is closed.
     */
    ConnectionErrors[ConnectionErrors["Closed"] = 1] = "Closed";
    /**
     * The connection got unsubscribed (i.e., disposed).
     */
    ConnectionErrors[ConnectionErrors["Unsubscribed"] = 2] = "Unsubscribed";
    /**
     * The connection is already in listening mode.
     */
    ConnectionErrors[ConnectionErrors["AlreadyListening"] = 3] = "AlreadyListening";
})(ConnectionErrors || (ConnectionErrors = {}));
export class ConnectionError extends Error {
    constructor(code, message) {
        super(message);
        this.code = code;
        Object.setPrototypeOf(this, ConnectionError.prototype);
    }
}
export function createConnection(transports, logger, strategy) {
    if (!logger) {
        logger = NullLogger;
    }
    return _createConnection(transports, logger, strategy);
}
var ConnectionState;
(function (ConnectionState) {
    ConnectionState[ConnectionState["New"] = 1] = "New";
    ConnectionState[ConnectionState["Listening"] = 2] = "Listening";
    ConnectionState[ConnectionState["Closed"] = 3] = "Closed";
    ConnectionState[ConnectionState["Unsubscribed"] = 4] = "Unsubscribed";
})(ConnectionState || (ConnectionState = {}));
function _createConnection(transports, logger, strategy) {
    let sequenceNumber = 0;
    let notificationSquenceNumber = 0;
    let unknownResponseSquenceNumber = 0;
    const version = '2.0';
    let starRequestHandler;
    const requestHandlers = Object.create(null);
    let starNotificationHandler;
    const notificationHandlers = Object.create(null);
    let timer = false;
    let messageQueue = new LinkedMap();
    let responsePromises = Object.create(null);
    let requestTokens = Object.create(null);
    let trace = Trace.Off;
    let tracer = noopTracer;
    let state = ConnectionState.New;
    const errorEmitter = new Emitter();
    const closeEmitter = new Emitter();
    const unhandledNotificationEmitter = new Emitter();
    const unsubscribeEmitter = new Emitter();
    function createRequestQueueKey(id) {
        return 'req-' + id.toString();
    }
    function createResponseQueueKey(id) {
        if (id === null) {
            return 'res-unknown-' + (++unknownResponseSquenceNumber).toString();
        }
        else {
            return 'res-' + id.toString();
        }
    }
    function createNotificationQueueKey() {
        return 'not-' + (++notificationSquenceNumber).toString();
    }
    function addMessageToQueue(queue, message) {
        if (isRequestMessage(message)) {
            queue.set(createRequestQueueKey(message.id), message);
        }
        else if (isResponseMessage(message)) {
            queue.set(createResponseQueueKey(message.id), message);
        }
        else {
            queue.set(createNotificationQueueKey(), message);
        }
    }
    function cancelUndispatched(_message) {
        return undefined;
    }
    function isListening() {
        return state === ConnectionState.Listening;
    }
    function isClosed() {
        return state === ConnectionState.Closed;
    }
    function isUnsubscribed() {
        return state === ConnectionState.Unsubscribed;
    }
    function closeHandler() {
        if (state === ConnectionState.New || state === ConnectionState.Listening) {
            state = ConnectionState.Closed;
            closeEmitter.fire(undefined);
        }
        // If the connection is unsubscribed don't sent close events.
    }
    function readErrorHandler(error) {
        errorEmitter.fire([error, undefined, undefined]);
    }
    function writeErrorHandler(data) {
        errorEmitter.fire(data);
    }
    transports.reader.onClose(closeHandler);
    transports.reader.onError(readErrorHandler);
    transports.writer.onClose(closeHandler);
    transports.writer.onError(writeErrorHandler);
    function triggerMessageQueue() {
        if (timer || messageQueue.size === 0) {
            return;
        }
        timer = true;
        setImmediateCompat(() => {
            timer = false;
            processMessageQueue();
        });
    }
    function processMessageQueue() {
        if (messageQueue.size === 0) {
            return;
        }
        const message = messageQueue.shift();
        try {
            if (isRequestMessage(message)) {
                handleRequest(message);
            }
            else if (isNotificationMessage(message)) {
                handleNotification(message);
            }
            else if (isResponseMessage(message)) {
                handleResponse(message);
            }
            else {
                handleInvalidMessage(message);
            }
        }
        finally {
            triggerMessageQueue();
        }
    }
    const callback = message => {
        try {
            // We have received a cancellation message. Check if the message is still in the queue and cancel it if
            // allowed to do so.
            if (isNotificationMessage(message) && message.method === CancelNotification.type) {
                const key = createRequestQueueKey(message.params.id);
                const toCancel = messageQueue.get(key);
                if (isRequestMessage(toCancel)) {
                    const response = strategy && strategy.cancelUndispatched
                        ? strategy.cancelUndispatched(toCancel, cancelUndispatched)
                        : cancelUndispatched(toCancel);
                    if (response && (response.error !== undefined || response.result !== undefined)) {
                        messageQueue.delete(key);
                        response.id = toCancel.id;
                        tracer.responseCanceled(response, toCancel, message);
                        transports.writer.write(response);
                        return;
                    }
                }
            }
            addMessageToQueue(messageQueue, message);
        }
        finally {
            triggerMessageQueue();
        }
    };
    function handleRequest(requestMessage) {
        if (isUnsubscribed()) {
            // we return here silently since we fired an event when the
            // connection got unsubscribed.
            return;
        }
        const startTime = Date.now();
        function reply(resultOrError) {
            const message = {
                jsonrpc: version,
                id: requestMessage.id,
            };
            if (resultOrError instanceof ResponseError) {
                message.error = resultOrError.toJSON();
            }
            else {
                message.result = resultOrError === undefined ? null : resultOrError;
            }
            tracer.responseSent(message, requestMessage, startTime);
            transports.writer.write(message);
        }
        function replyError(error) {
            const message = {
                jsonrpc: version,
                id: requestMessage.id,
                error: error.toJSON(),
            };
            tracer.responseSent(message, requestMessage, startTime);
            transports.writer.write(message);
        }
        function replySuccess(result) {
            // The JSON RPC defines that a response must either have a result or an error
            // So we can't treat undefined as a valid response result.
            if (result === undefined) {
                result = null;
            }
            const message = {
                jsonrpc: version,
                id: requestMessage.id,
                result,
            };
            tracer.responseSent(message, requestMessage, startTime);
            transports.writer.write(message);
        }
        tracer.requestReceived(requestMessage);
        const element = requestHandlers[requestMessage.method];
        const requestHandler = element && element.handler;
        if (requestHandler || starRequestHandler) {
            const cancellationSource = new CancellationTokenSource();
            const tokenKey = String(requestMessage.id);
            requestTokens[tokenKey] = cancellationSource;
            try {
                const params = requestMessage.params !== undefined ? requestMessage.params : null;
                const handlerResult = requestHandler
                    ? requestHandler(params, cancellationSource.token)
                    : starRequestHandler(requestMessage.method, params, cancellationSource.token);
                const promise = handlerResult;
                if (!handlerResult) {
                    delete requestTokens[tokenKey];
                    replySuccess(handlerResult);
                }
                else if (promise.then) {
                    promise.then((resultOrError) => {
                        delete requestTokens[tokenKey];
                        reply(resultOrError);
                    }, error => {
                        delete requestTokens[tokenKey];
                        if (error instanceof ResponseError) {
                            replyError(error);
                        }
                        else if (error && typeof error.message === 'string') {
                            replyError(new ResponseError(ErrorCodes.InternalError, error.message, Object.assign({ stack: error.stack }, error)));
                        }
                        else {
                            replyError(new ResponseError(ErrorCodes.InternalError, `Request ${requestMessage.method} failed unexpectedly without providing any details.`));
                        }
                    });
                }
                else {
                    delete requestTokens[tokenKey];
                    reply(handlerResult);
                }
            }
            catch (error) {
                delete requestTokens[tokenKey];
                if (error instanceof ResponseError) {
                    reply(error);
                }
                else if (error && typeof error.message === 'string') {
                    replyError(new ResponseError(ErrorCodes.InternalError, error.message, Object.assign({ stack: error.stack }, error)));
                }
                else {
                    replyError(new ResponseError(ErrorCodes.InternalError, `Request ${requestMessage.method} failed unexpectedly without providing any details.`));
                }
            }
        }
        else {
            replyError(new ResponseError(ErrorCodes.MethodNotFound, `Unhandled method ${requestMessage.method}`));
        }
    }
    function handleResponse(responseMessage) {
        if (isUnsubscribed()) {
            // See handle request.
            return;
        }
        if (responseMessage.id === null) {
            if (responseMessage.error) {
                logger.error(`Received response message without id: Error is: \n${JSON.stringify(responseMessage.error, undefined, 4)}`);
            }
            else {
                logger.error(`Received response message without id. No further error information provided.`);
            }
        }
        else {
            const key = String(responseMessage.id);
            const responsePromise = responsePromises[key];
            if (responsePromise) {
                tracer.responseReceived(responseMessage, responsePromise.request || responsePromise.method, responsePromise.timerStart);
                delete responsePromises[key];
                try {
                    if (responseMessage.error) {
                        const error = responseMessage.error;
                        responsePromise.reject(new ResponseError(error.code, error.message, error.data));
                    }
                    else if (responseMessage.result !== undefined) {
                        responsePromise.resolve(responseMessage.result);
                    }
                    else {
                        throw new Error('Should never happen.');
                    }
                }
                catch (error) {
                    if (error.message) {
                        logger.error(`Response handler '${responsePromise.method}' failed with message: ${error.message}`);
                    }
                    else {
                        logger.error(`Response handler '${responsePromise.method}' failed unexpectedly.`);
                    }
                }
            }
            else {
                tracer.unknownResponseReceived(responseMessage);
            }
        }
    }
    function handleNotification(message) {
        if (isUnsubscribed()) {
            // See handle request.
            return;
        }
        let notificationHandler;
        if (message.method === CancelNotification.type) {
            notificationHandler = (params) => {
                const id = params.id;
                const source = requestTokens[String(id)];
                if (source) {
                    source.cancel();
                }
            };
        }
        else {
            const element = notificationHandlers[message.method];
            if (element) {
                notificationHandler = element.handler;
            }
        }
        if (notificationHandler || starNotificationHandler) {
            try {
                tracer.notificationReceived(message);
                notificationHandler
                    ? notificationHandler(message.params)
                    : starNotificationHandler(message.method, message.params);
            }
            catch (error) {
                if (error.message) {
                    logger.error(`Notification handler '${message.method}' failed with message: ${error.message}`);
                }
                else {
                    logger.error(`Notification handler '${message.method}' failed unexpectedly.`);
                }
            }
        }
        else {
            unhandledNotificationEmitter.fire(message);
        }
    }
    function handleInvalidMessage(message) {
        if (!message) {
            logger.error('Received empty message.');
            return;
        }
        logger.error(`Received message which is neither a response nor a notification message:\n${JSON.stringify(message, null, 4)}`);
        // Test whether we find an id to reject the promise
        const responseMessage = message;
        if (typeof responseMessage.id === 'string' || typeof responseMessage.id === 'number') {
            const key = String(responseMessage.id);
            const responseHandler = responsePromises[key];
            if (responseHandler) {
                responseHandler.reject(new Error('The received response has neither a result nor an error property.'));
            }
        }
    }
    function throwIfClosedOrUnsubscribed() {
        if (isClosed()) {
            throw new ConnectionError(ConnectionErrors.Closed, 'Connection is closed.');
        }
        if (isUnsubscribed()) {
            throw new ConnectionError(ConnectionErrors.Unsubscribed, 'Connection is unsubscribed.');
        }
    }
    function throwIfListening() {
        if (isListening()) {
            throw new ConnectionError(ConnectionErrors.AlreadyListening, 'Connection is already listening');
        }
    }
    function throwIfNotListening() {
        if (!isListening()) {
            throw new Error('Call listen() first.');
        }
    }
    const connection = {
        sendNotification: (method, params) => {
            throwIfClosedOrUnsubscribed();
            const notificationMessage = {
                jsonrpc: version,
                method,
                params,
            };
            tracer.notificationSent(notificationMessage);
            transports.writer.write(notificationMessage);
        },
        onNotification: (type, handler) => {
            throwIfClosedOrUnsubscribed();
            if (typeof type === 'function') {
                starNotificationHandler = type;
            }
            else if (handler) {
                notificationHandlers[type] = { type: undefined, handler };
            }
        },
        sendRequest: (method, params, token) => {
            throwIfClosedOrUnsubscribed();
            throwIfNotListening();
            token = CancellationToken.is(token) ? token : undefined;
            const id = sequenceNumber++;
            const result = new Promise((resolve, reject) => {
                const requestMessage = {
                    jsonrpc: version,
                    id,
                    method,
                    params,
                };
                let responsePromise = {
                    method,
                    request: trace === Trace.Verbose ? requestMessage : undefined,
                    timerStart: Date.now(),
                    resolve,
                    reject,
                };
                tracer.requestSent(requestMessage);
                try {
                    transports.writer.write(requestMessage);
                }
                catch (e) {
                    // Writing the message failed. So we need to reject the promise.
                    responsePromise.reject(new ResponseError(ErrorCodes.MessageWriteError, e.message ? e.message : 'Unknown reason'));
                    responsePromise = null;
                }
                if (responsePromise) {
                    responsePromises[String(id)] = responsePromise;
                }
            });
            if (token) {
                token.onCancellationRequested(() => {
                    connection.sendNotification(CancelNotification.type, { id });
                });
            }
            return result;
        },
        onRequest: (type, handler) => {
            throwIfClosedOrUnsubscribed();
            if (typeof type === 'function') {
                starRequestHandler = type;
            }
            else if (handler) {
                requestHandlers[type] = { type: undefined, handler };
            }
        },
        trace: (value, _tracer, sendNotification = false) => {
            trace = value;
            if (trace === Trace.Off) {
                tracer = noopTracer;
            }
            else {
                tracer = _tracer;
            }
        },
        onError: errorEmitter.event,
        onClose: closeEmitter.event,
        onUnhandledNotification: unhandledNotificationEmitter.event,
        onUnsubscribe: unsubscribeEmitter.event,
        unsubscribe: () => {
            if (isUnsubscribed()) {
                return;
            }
            state = ConnectionState.Unsubscribed;
            unsubscribeEmitter.fire(undefined);
            for (const key of Object.keys(responsePromises)) {
                responsePromises[key].reject(new ConnectionError(ConnectionErrors.Unsubscribed, `The underlying JSON-RPC connection got unsubscribed while responding to this ${responsePromises[key].method} request.`));
            }
            responsePromises = Object.create(null);
            requestTokens = Object.create(null);
            messageQueue = new LinkedMap();
            transports.writer.unsubscribe();
            transports.reader.unsubscribe();
        },
        listen: () => {
            throwIfClosedOrUnsubscribed();
            throwIfListening();
            state = ConnectionState.Listening;
            transports.reader.listen(callback);
        },
    };
    return connection;
}
/** Support browser and node environments without needing a transpiler. */
function setImmediateCompat(f) {
    if (typeof setImmediate !== 'undefined') {
        setImmediate(f);
        return;
    }
    setTimeout(f, 0);
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY29ubmVjdGlvbi5qcyIsInNvdXJjZVJvb3QiOiJzcmMvIiwic291cmNlcyI6WyJwcm90b2NvbC9qc29ucnBjMi9jb25uZWN0aW9uLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUNBLE9BQU8sRUFBRSxrQkFBa0IsRUFBZ0IsTUFBTSxVQUFVLENBQUE7QUFDM0QsT0FBTyxFQUFFLGlCQUFpQixFQUFFLHVCQUF1QixFQUFFLE1BQU0sVUFBVSxDQUFBO0FBRXJFLE9BQU8sRUFBRSxPQUFPLEVBQVMsTUFBTSxVQUFVLENBQUE7QUFPekMsT0FBTyxFQUFFLFNBQVMsRUFBRSxNQUFNLGFBQWEsQ0FBQTtBQUN2QyxPQUFPLEVBQ0gsVUFBVSxFQUNWLHFCQUFxQixFQUNyQixnQkFBZ0IsRUFDaEIsaUJBQWlCLEVBSWpCLGFBQWEsR0FFaEIsTUFBTSxZQUFZLENBQUE7QUFDbkIsT0FBTyxFQUFFLFVBQVUsRUFBRSxLQUFLLEVBQVUsTUFBTSxTQUFTLENBQUE7QUFZbkQsTUFBTSxVQUFVLEdBQVcsTUFBTSxDQUFDLE1BQU0sQ0FBQztJQUNyQyxLQUFLLEVBQUUsR0FBRyxFQUFFO1FBQ1IsVUFBVTtJQUNkLENBQUM7SUFDRCxJQUFJLEVBQUUsR0FBRyxFQUFFO1FBQ1AsVUFBVTtJQUNkLENBQUM7SUFDRCxJQUFJLEVBQUUsR0FBRyxFQUFFO1FBQ1AsVUFBVTtJQUNkLENBQUM7SUFDRCxHQUFHLEVBQUUsR0FBRyxFQUFFO1FBQ04sVUFBVTtJQUNkLENBQUM7Q0FDSixDQUFDLENBQUE7QUFFRixNQUFNLENBQU4sSUFBWSxnQkFhWDtBQWJELFdBQVksZ0JBQWdCO0lBQ3hCOztPQUVHO0lBQ0gsMkRBQVUsQ0FBQTtJQUNWOztPQUVHO0lBQ0gsdUVBQWdCLENBQUE7SUFDaEI7O09BRUc7SUFDSCwrRUFBb0IsQ0FBQTtBQUN4QixDQUFDLEVBYlcsZ0JBQWdCLEtBQWhCLGdCQUFnQixRQWEzQjtBQUVELE1BQU0sT0FBTyxlQUFnQixTQUFRLEtBQUs7SUFHdEMsWUFBWSxJQUFzQixFQUFFLE9BQWU7UUFDL0MsS0FBSyxDQUFDLE9BQU8sQ0FBQyxDQUFBO1FBQ2QsSUFBSSxDQUFDLElBQUksR0FBRyxJQUFJLENBQUE7UUFDaEIsTUFBTSxDQUFDLGNBQWMsQ0FBQyxJQUFJLEVBQUUsZUFBZSxDQUFDLFNBQVMsQ0FBQyxDQUFBO0lBQzFELENBQUM7Q0FDSjtBQThCRCxNQUFNLFVBQVUsZ0JBQWdCLENBQzVCLFVBQTZCLEVBQzdCLE1BQWUsRUFDZixRQUE2QjtJQUU3QixJQUFJLENBQUMsTUFBTSxFQUFFO1FBQ1QsTUFBTSxHQUFHLFVBQVUsQ0FBQTtLQUN0QjtJQUNELE9BQU8saUJBQWlCLENBQUMsVUFBVSxFQUFFLE1BQU0sRUFBRSxRQUFRLENBQUMsQ0FBQTtBQUMxRCxDQUFDO0FBZ0JELElBQUssZUFLSjtBQUxELFdBQUssZUFBZTtJQUNoQixtREFBTyxDQUFBO0lBQ1AsK0RBQWEsQ0FBQTtJQUNiLHlEQUFVLENBQUE7SUFDVixxRUFBZ0IsQ0FBQTtBQUNwQixDQUFDLEVBTEksZUFBZSxLQUFmLGVBQWUsUUFLbkI7QUFZRCxTQUFTLGlCQUFpQixDQUFDLFVBQTZCLEVBQUUsTUFBYyxFQUFFLFFBQTZCO0lBQ25HLElBQUksY0FBYyxHQUFHLENBQUMsQ0FBQTtJQUN0QixJQUFJLHlCQUF5QixHQUFHLENBQUMsQ0FBQTtJQUNqQyxJQUFJLDRCQUE0QixHQUFHLENBQUMsQ0FBQTtJQUNwQyxNQUFNLE9BQU8sR0FBRyxLQUFLLENBQUE7SUFFckIsSUFBSSxrQkFBa0QsQ0FBQTtJQUN0RCxNQUFNLGVBQWUsR0FBMEQsTUFBTSxDQUFDLE1BQU0sQ0FBQyxJQUFJLENBQUMsQ0FBQTtJQUNsRyxJQUFJLHVCQUE0RCxDQUFBO0lBQ2hFLE1BQU0sb0JBQW9CLEdBQStELE1BQU0sQ0FBQyxNQUFNLENBQUMsSUFBSSxDQUFDLENBQUE7SUFFNUcsSUFBSSxLQUFLLEdBQUcsS0FBSyxDQUFBO0lBQ2pCLElBQUksWUFBWSxHQUFpQixJQUFJLFNBQVMsRUFBbUIsQ0FBQTtJQUNqRSxJQUFJLGdCQUFnQixHQUF3QyxNQUFNLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxDQUFBO0lBQy9FLElBQUksYUFBYSxHQUE4QyxNQUFNLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxDQUFBO0lBRWxGLElBQUksS0FBSyxHQUFVLEtBQUssQ0FBQyxHQUFHLENBQUE7SUFDNUIsSUFBSSxNQUFNLEdBQVcsVUFBVSxDQUFBO0lBRS9CLElBQUksS0FBSyxHQUFvQixlQUFlLENBQUMsR0FBRyxDQUFBO0lBQ2hELE1BQU0sWUFBWSxHQUFHLElBQUksT0FBTyxFQUFvRCxDQUFBO0lBQ3BGLE1BQU0sWUFBWSxHQUFrQixJQUFJLE9BQU8sRUFBUSxDQUFBO0lBQ3ZELE1BQU0sNEJBQTRCLEdBQWlDLElBQUksT0FBTyxFQUF1QixDQUFBO0lBRXJHLE1BQU0sa0JBQWtCLEdBQWtCLElBQUksT0FBTyxFQUFRLENBQUE7SUFFN0QsU0FBUyxxQkFBcUIsQ0FBQyxFQUFtQjtRQUM5QyxPQUFPLE1BQU0sR0FBRyxFQUFFLENBQUMsUUFBUSxFQUFFLENBQUE7SUFDakMsQ0FBQztJQUVELFNBQVMsc0JBQXNCLENBQUMsRUFBMEI7UUFDdEQsSUFBSSxFQUFFLEtBQUssSUFBSSxFQUFFO1lBQ2IsT0FBTyxjQUFjLEdBQUcsQ0FBQyxFQUFFLDRCQUE0QixDQUFDLENBQUMsUUFBUSxFQUFFLENBQUE7U0FDdEU7YUFBTTtZQUNILE9BQU8sTUFBTSxHQUFHLEVBQUUsQ0FBQyxRQUFRLEVBQUUsQ0FBQTtTQUNoQztJQUNMLENBQUM7SUFFRCxTQUFTLDBCQUEwQjtRQUMvQixPQUFPLE1BQU0sR0FBRyxDQUFDLEVBQUUseUJBQXlCLENBQUMsQ0FBQyxRQUFRLEVBQUUsQ0FBQTtJQUM1RCxDQUFDO0lBRUQsU0FBUyxpQkFBaUIsQ0FBQyxLQUFtQixFQUFFLE9BQWdCO1FBQzVELElBQUksZ0JBQWdCLENBQUMsT0FBTyxDQUFDLEVBQUU7WUFDM0IsS0FBSyxDQUFDLEdBQUcsQ0FBQyxxQkFBcUIsQ0FBQyxPQUFPLENBQUMsRUFBRSxDQUFDLEVBQUUsT0FBTyxDQUFDLENBQUE7U0FDeEQ7YUFBTSxJQUFJLGlCQUFpQixDQUFDLE9BQU8sQ0FBQyxFQUFFO1lBQ25DLEtBQUssQ0FBQyxHQUFHLENBQUMsc0JBQXNCLENBQUMsT0FBTyxDQUFDLEVBQUUsQ0FBQyxFQUFFLE9BQU8sQ0FBQyxDQUFBO1NBQ3pEO2FBQU07WUFDSCxLQUFLLENBQUMsR0FBRyxDQUFDLDBCQUEwQixFQUFFLEVBQUUsT0FBTyxDQUFDLENBQUE7U0FDbkQ7SUFDTCxDQUFDO0lBRUQsU0FBUyxrQkFBa0IsQ0FBQyxRQUFpQjtRQUN6QyxPQUFPLFNBQVMsQ0FBQTtJQUNwQixDQUFDO0lBRUQsU0FBUyxXQUFXO1FBQ2hCLE9BQU8sS0FBSyxLQUFLLGVBQWUsQ0FBQyxTQUFTLENBQUE7SUFDOUMsQ0FBQztJQUVELFNBQVMsUUFBUTtRQUNiLE9BQU8sS0FBSyxLQUFLLGVBQWUsQ0FBQyxNQUFNLENBQUE7SUFDM0MsQ0FBQztJQUVELFNBQVMsY0FBYztRQUNuQixPQUFPLEtBQUssS0FBSyxlQUFlLENBQUMsWUFBWSxDQUFBO0lBQ2pELENBQUM7SUFFRCxTQUFTLFlBQVk7UUFDakIsSUFBSSxLQUFLLEtBQUssZUFBZSxDQUFDLEdBQUcsSUFBSSxLQUFLLEtBQUssZUFBZSxDQUFDLFNBQVMsRUFBRTtZQUN0RSxLQUFLLEdBQUcsZUFBZSxDQUFDLE1BQU0sQ0FBQTtZQUM5QixZQUFZLENBQUMsSUFBSSxDQUFDLFNBQVMsQ0FBQyxDQUFBO1NBQy9CO1FBQ0QsNkRBQTZEO0lBQ2pFLENBQUM7SUFFRCxTQUFTLGdCQUFnQixDQUFDLEtBQVk7UUFDbEMsWUFBWSxDQUFDLElBQUksQ0FBQyxDQUFDLEtBQUssRUFBRSxTQUFTLEVBQUUsU0FBUyxDQUFDLENBQUMsQ0FBQTtJQUNwRCxDQUFDO0lBRUQsU0FBUyxpQkFBaUIsQ0FBQyxJQUFzRDtRQUM3RSxZQUFZLENBQUMsSUFBSSxDQUFDLElBQUksQ0FBQyxDQUFBO0lBQzNCLENBQUM7SUFFRCxVQUFVLENBQUMsTUFBTSxDQUFDLE9BQU8sQ0FBQyxZQUFZLENBQUMsQ0FBQTtJQUN2QyxVQUFVLENBQUMsTUFBTSxDQUFDLE9BQU8sQ0FBQyxnQkFBZ0IsQ0FBQyxDQUFBO0lBRTNDLFVBQVUsQ0FBQyxNQUFNLENBQUMsT0FBTyxDQUFDLFlBQVksQ0FBQyxDQUFBO0lBQ3ZDLFVBQVUsQ0FBQyxNQUFNLENBQUMsT0FBTyxDQUFDLGlCQUFpQixDQUFDLENBQUE7SUFFNUMsU0FBUyxtQkFBbUI7UUFDeEIsSUFBSSxLQUFLLElBQUksWUFBWSxDQUFDLElBQUksS0FBSyxDQUFDLEVBQUU7WUFDbEMsT0FBTTtTQUNUO1FBQ0QsS0FBSyxHQUFHLElBQUksQ0FBQTtRQUNaLGtCQUFrQixDQUFDLEdBQUcsRUFBRTtZQUNwQixLQUFLLEdBQUcsS0FBSyxDQUFBO1lBQ2IsbUJBQW1CLEVBQUUsQ0FBQTtRQUN6QixDQUFDLENBQUMsQ0FBQTtJQUNOLENBQUM7SUFFRCxTQUFTLG1CQUFtQjtRQUN4QixJQUFJLFlBQVksQ0FBQyxJQUFJLEtBQUssQ0FBQyxFQUFFO1lBQ3pCLE9BQU07U0FDVDtRQUNELE1BQU0sT0FBTyxHQUFHLFlBQVksQ0FBQyxLQUFLLEVBQUcsQ0FBQTtRQUNyQyxJQUFJO1lBQ0EsSUFBSSxnQkFBZ0IsQ0FBQyxPQUFPLENBQUMsRUFBRTtnQkFDM0IsYUFBYSxDQUFDLE9BQU8sQ0FBQyxDQUFBO2FBQ3pCO2lCQUFNLElBQUkscUJBQXFCLENBQUMsT0FBTyxDQUFDLEVBQUU7Z0JBQ3ZDLGtCQUFrQixDQUFDLE9BQU8sQ0FBQyxDQUFBO2FBQzlCO2lCQUFNLElBQUksaUJBQWlCLENBQUMsT0FBTyxDQUFDLEVBQUU7Z0JBQ25DLGNBQWMsQ0FBQyxPQUFPLENBQUMsQ0FBQTthQUMxQjtpQkFBTTtnQkFDSCxvQkFBb0IsQ0FBQyxPQUFPLENBQUMsQ0FBQTthQUNoQztTQUNKO2dCQUFTO1lBQ04sbUJBQW1CLEVBQUUsQ0FBQTtTQUN4QjtJQUNMLENBQUM7SUFFRCxNQUFNLFFBQVEsR0FBaUIsT0FBTyxDQUFDLEVBQUU7UUFDckMsSUFBSTtZQUNBLHVHQUF1RztZQUN2RyxvQkFBb0I7WUFDcEIsSUFBSSxxQkFBcUIsQ0FBQyxPQUFPLENBQUMsSUFBSSxPQUFPLENBQUMsTUFBTSxLQUFLLGtCQUFrQixDQUFDLElBQUksRUFBRTtnQkFDOUUsTUFBTSxHQUFHLEdBQUcscUJBQXFCLENBQUUsT0FBTyxDQUFDLE1BQXVCLENBQUMsRUFBRSxDQUFDLENBQUE7Z0JBQ3RFLE1BQU0sUUFBUSxHQUFHLFlBQVksQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLENBQUE7Z0JBQ3RDLElBQUksZ0JBQWdCLENBQUMsUUFBUSxDQUFDLEVBQUU7b0JBQzVCLE1BQU0sUUFBUSxHQUNWLFFBQVEsSUFBSSxRQUFRLENBQUMsa0JBQWtCO3dCQUNuQyxDQUFDLENBQUMsUUFBUSxDQUFDLGtCQUFrQixDQUFDLFFBQVEsRUFBRSxrQkFBa0IsQ0FBQzt3QkFDM0QsQ0FBQyxDQUFDLGtCQUFrQixDQUFDLFFBQVEsQ0FBQyxDQUFBO29CQUN0QyxJQUFJLFFBQVEsSUFBSSxDQUFDLFFBQVEsQ0FBQyxLQUFLLEtBQUssU0FBUyxJQUFJLFFBQVEsQ0FBQyxNQUFNLEtBQUssU0FBUyxDQUFDLEVBQUU7d0JBQzdFLFlBQVksQ0FBQyxNQUFNLENBQUMsR0FBRyxDQUFDLENBQUE7d0JBQ3hCLFFBQVEsQ0FBQyxFQUFFLEdBQUcsUUFBUSxDQUFDLEVBQUUsQ0FBQTt3QkFDekIsTUFBTSxDQUFDLGdCQUFnQixDQUFDLFFBQVEsRUFBRSxRQUFRLEVBQUUsT0FBTyxDQUFDLENBQUE7d0JBQ3BELFVBQVUsQ0FBQyxNQUFNLENBQUMsS0FBSyxDQUFDLFFBQVEsQ0FBQyxDQUFBO3dCQUNqQyxPQUFNO3FCQUNUO2lCQUNKO2FBQ0o7WUFDRCxpQkFBaUIsQ0FBQyxZQUFZLEVBQUUsT0FBTyxDQUFDLENBQUE7U0FDM0M7Z0JBQVM7WUFDTixtQkFBbUIsRUFBRSxDQUFBO1NBQ3hCO0lBQ0wsQ0FBQyxDQUFBO0lBRUQsU0FBUyxhQUFhLENBQUMsY0FBOEI7UUFDakQsSUFBSSxjQUFjLEVBQUUsRUFBRTtZQUNsQiwyREFBMkQ7WUFDM0QsK0JBQStCO1lBQy9CLE9BQU07U0FDVDtRQUVELE1BQU0sU0FBUyxHQUFHLElBQUksQ0FBQyxHQUFHLEVBQUUsQ0FBQTtRQUU1QixTQUFTLEtBQUssQ0FBQyxhQUF1QztZQUNsRCxNQUFNLE9BQU8sR0FBb0I7Z0JBQzdCLE9BQU8sRUFBRSxPQUFPO2dCQUNoQixFQUFFLEVBQUUsY0FBYyxDQUFDLEVBQUU7YUFDeEIsQ0FBQTtZQUNELElBQUksYUFBYSxZQUFZLGFBQWEsRUFBRTtnQkFDeEMsT0FBTyxDQUFDLEtBQUssR0FBSSxhQUFvQyxDQUFDLE1BQU0sRUFBRSxDQUFBO2FBQ2pFO2lCQUFNO2dCQUNILE9BQU8sQ0FBQyxNQUFNLEdBQUcsYUFBYSxLQUFLLFNBQVMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxhQUFhLENBQUE7YUFDdEU7WUFDRCxNQUFNLENBQUMsWUFBWSxDQUFDLE9BQU8sRUFBRSxjQUFjLEVBQUUsU0FBUyxDQUFDLENBQUE7WUFDdkQsVUFBVSxDQUFDLE1BQU0sQ0FBQyxLQUFLLENBQUMsT0FBTyxDQUFDLENBQUE7UUFDcEMsQ0FBQztRQUNELFNBQVMsVUFBVSxDQUFDLEtBQXlCO1lBQ3pDLE1BQU0sT0FBTyxHQUFvQjtnQkFDN0IsT0FBTyxFQUFFLE9BQU87Z0JBQ2hCLEVBQUUsRUFBRSxjQUFjLENBQUMsRUFBRTtnQkFDckIsS0FBSyxFQUFFLEtBQUssQ0FBQyxNQUFNLEVBQUU7YUFDeEIsQ0FBQTtZQUNELE1BQU0sQ0FBQyxZQUFZLENBQUMsT0FBTyxFQUFFLGNBQWMsRUFBRSxTQUFTLENBQUMsQ0FBQTtZQUN2RCxVQUFVLENBQUMsTUFBTSxDQUFDLEtBQUssQ0FBQyxPQUFPLENBQUMsQ0FBQTtRQUNwQyxDQUFDO1FBQ0QsU0FBUyxZQUFZLENBQUMsTUFBVztZQUM3Qiw2RUFBNkU7WUFDN0UsMERBQTBEO1lBQzFELElBQUksTUFBTSxLQUFLLFNBQVMsRUFBRTtnQkFDdEIsTUFBTSxHQUFHLElBQUksQ0FBQTthQUNoQjtZQUNELE1BQU0sT0FBTyxHQUFvQjtnQkFDN0IsT0FBTyxFQUFFLE9BQU87Z0JBQ2hCLEVBQUUsRUFBRSxjQUFjLENBQUMsRUFBRTtnQkFDckIsTUFBTTthQUNULENBQUE7WUFDRCxNQUFNLENBQUMsWUFBWSxDQUFDLE9BQU8sRUFBRSxjQUFjLEVBQUUsU0FBUyxDQUFDLENBQUE7WUFDdkQsVUFBVSxDQUFDLE1BQU0sQ0FBQyxLQUFLLENBQUMsT0FBTyxDQUFDLENBQUE7UUFDcEMsQ0FBQztRQUVELE1BQU0sQ0FBQyxlQUFlLENBQUMsY0FBYyxDQUFDLENBQUE7UUFFdEMsTUFBTSxPQUFPLEdBQUcsZUFBZSxDQUFDLGNBQWMsQ0FBQyxNQUFNLENBQUMsQ0FBQTtRQUN0RCxNQUFNLGNBQWMsR0FBZ0QsT0FBTyxJQUFJLE9BQU8sQ0FBQyxPQUFPLENBQUE7UUFDOUYsSUFBSSxjQUFjLElBQUksa0JBQWtCLEVBQUU7WUFDdEMsTUFBTSxrQkFBa0IsR0FBRyxJQUFJLHVCQUF1QixFQUFFLENBQUE7WUFDeEQsTUFBTSxRQUFRLEdBQUcsTUFBTSxDQUFDLGNBQWMsQ0FBQyxFQUFFLENBQUMsQ0FBQTtZQUMxQyxhQUFhLENBQUMsUUFBUSxDQUFDLEdBQUcsa0JBQWtCLENBQUE7WUFDNUMsSUFBSTtnQkFDQSxNQUFNLE1BQU0sR0FBRyxjQUFjLENBQUMsTUFBTSxLQUFLLFNBQVMsQ0FBQyxDQUFDLENBQUMsY0FBYyxDQUFDLE1BQU0sQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFBO2dCQUNqRixNQUFNLGFBQWEsR0FBRyxjQUFjO29CQUNoQyxDQUFDLENBQUMsY0FBYyxDQUFDLE1BQU0sRUFBRSxrQkFBa0IsQ0FBQyxLQUFLLENBQUM7b0JBQ2xELENBQUMsQ0FBQyxrQkFBbUIsQ0FBQyxjQUFjLENBQUMsTUFBTSxFQUFFLE1BQU0sRUFBRSxrQkFBa0IsQ0FBQyxLQUFLLENBQUMsQ0FBQTtnQkFFbEYsTUFBTSxPQUFPLEdBQUcsYUFBa0QsQ0FBQTtnQkFDbEUsSUFBSSxDQUFDLGFBQWEsRUFBRTtvQkFDaEIsT0FBTyxhQUFhLENBQUMsUUFBUSxDQUFDLENBQUE7b0JBQzlCLFlBQVksQ0FBQyxhQUFhLENBQUMsQ0FBQTtpQkFDOUI7cUJBQU0sSUFBSSxPQUFPLENBQUMsSUFBSSxFQUFFO29CQUNyQixPQUFPLENBQUMsSUFBSSxDQUNSLENBQUMsYUFBYSxFQUE0QixFQUFFO3dCQUN4QyxPQUFPLGFBQWEsQ0FBQyxRQUFRLENBQUMsQ0FBQTt3QkFDOUIsS0FBSyxDQUFDLGFBQWEsQ0FBQyxDQUFBO29CQUN4QixDQUFDLEVBQ0QsS0FBSyxDQUFDLEVBQUU7d0JBQ0osT0FBTyxhQUFhLENBQUMsUUFBUSxDQUFDLENBQUE7d0JBQzlCLElBQUksS0FBSyxZQUFZLGFBQWEsRUFBRTs0QkFDaEMsVUFBVSxDQUFDLEtBQTJCLENBQUMsQ0FBQTt5QkFDMUM7NkJBQU0sSUFBSSxLQUFLLElBQUksT0FBTyxLQUFLLENBQUMsT0FBTyxLQUFLLFFBQVEsRUFBRTs0QkFDbkQsVUFBVSxDQUNOLElBQUksYUFBYSxDQUFPLFVBQVUsQ0FBQyxhQUFhLEVBQUUsS0FBSyxDQUFDLE9BQU8sa0JBQzNELEtBQUssRUFBRSxLQUFLLENBQUMsS0FBSyxJQUNmLEtBQUssRUFDVixDQUNMLENBQUE7eUJBQ0o7NkJBQU07NEJBQ0gsVUFBVSxDQUNOLElBQUksYUFBYSxDQUNiLFVBQVUsQ0FBQyxhQUFhLEVBQ3hCLFdBQ0ksY0FBYyxDQUFDLE1BQ25CLHFEQUFxRCxDQUN4RCxDQUNKLENBQUE7eUJBQ0o7b0JBQ0wsQ0FBQyxDQUNKLENBQUE7aUJBQ0o7cUJBQU07b0JBQ0gsT0FBTyxhQUFhLENBQUMsUUFBUSxDQUFDLENBQUE7b0JBQzlCLEtBQUssQ0FBQyxhQUFhLENBQUMsQ0FBQTtpQkFDdkI7YUFDSjtZQUFDLE9BQU8sS0FBSyxFQUFFO2dCQUNaLE9BQU8sYUFBYSxDQUFDLFFBQVEsQ0FBQyxDQUFBO2dCQUM5QixJQUFJLEtBQUssWUFBWSxhQUFhLEVBQUU7b0JBQ2hDLEtBQUssQ0FBQyxLQUEyQixDQUFDLENBQUE7aUJBQ3JDO3FCQUFNLElBQUksS0FBSyxJQUFJLE9BQU8sS0FBSyxDQUFDLE9BQU8sS0FBSyxRQUFRLEVBQUU7b0JBQ25ELFVBQVUsQ0FDTixJQUFJLGFBQWEsQ0FBTyxVQUFVLENBQUMsYUFBYSxFQUFFLEtBQUssQ0FBQyxPQUFPLGtCQUMzRCxLQUFLLEVBQUUsS0FBSyxDQUFDLEtBQUssSUFDZixLQUFLLEVBQ1YsQ0FDTCxDQUFBO2lCQUNKO3FCQUFNO29CQUNILFVBQVUsQ0FDTixJQUFJLGFBQWEsQ0FDYixVQUFVLENBQUMsYUFBYSxFQUN4QixXQUFXLGNBQWMsQ0FBQyxNQUFNLHFEQUFxRCxDQUN4RixDQUNKLENBQUE7aUJBQ0o7YUFDSjtTQUNKO2FBQU07WUFDSCxVQUFVLENBQUMsSUFBSSxhQUFhLENBQU8sVUFBVSxDQUFDLGNBQWMsRUFBRSxvQkFBb0IsY0FBYyxDQUFDLE1BQU0sRUFBRSxDQUFDLENBQUMsQ0FBQTtTQUM5RztJQUNMLENBQUM7SUFFRCxTQUFTLGNBQWMsQ0FBQyxlQUFnQztRQUNwRCxJQUFJLGNBQWMsRUFBRSxFQUFFO1lBQ2xCLHNCQUFzQjtZQUN0QixPQUFNO1NBQ1Q7UUFFRCxJQUFJLGVBQWUsQ0FBQyxFQUFFLEtBQUssSUFBSSxFQUFFO1lBQzdCLElBQUksZUFBZSxDQUFDLEtBQUssRUFBRTtnQkFDdkIsTUFBTSxDQUFDLEtBQUssQ0FDUixxREFBcUQsSUFBSSxDQUFDLFNBQVMsQ0FDL0QsZUFBZSxDQUFDLEtBQUssRUFDckIsU0FBUyxFQUNULENBQUMsQ0FDSixFQUFFLENBQ04sQ0FBQTthQUNKO2lCQUFNO2dCQUNILE1BQU0sQ0FBQyxLQUFLLENBQUMsOEVBQThFLENBQUMsQ0FBQTthQUMvRjtTQUNKO2FBQU07WUFDSCxNQUFNLEdBQUcsR0FBRyxNQUFNLENBQUMsZUFBZSxDQUFDLEVBQUUsQ0FBQyxDQUFBO1lBQ3RDLE1BQU0sZUFBZSxHQUFHLGdCQUFnQixDQUFDLEdBQUcsQ0FBQyxDQUFBO1lBQzdDLElBQUksZUFBZSxFQUFFO2dCQUNqQixNQUFNLENBQUMsZ0JBQWdCLENBQ25CLGVBQWUsRUFDZixlQUFlLENBQUMsT0FBTyxJQUFJLGVBQWUsQ0FBQyxNQUFNLEVBQ2pELGVBQWUsQ0FBQyxVQUFVLENBQzdCLENBQUE7Z0JBQ0QsT0FBTyxnQkFBZ0IsQ0FBQyxHQUFHLENBQUMsQ0FBQTtnQkFDNUIsSUFBSTtvQkFDQSxJQUFJLGVBQWUsQ0FBQyxLQUFLLEVBQUU7d0JBQ3ZCLE1BQU0sS0FBSyxHQUFHLGVBQWUsQ0FBQyxLQUFLLENBQUE7d0JBQ25DLGVBQWUsQ0FBQyxNQUFNLENBQUMsSUFBSSxhQUFhLENBQUMsS0FBSyxDQUFDLElBQUksRUFBRSxLQUFLLENBQUMsT0FBTyxFQUFFLEtBQUssQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFBO3FCQUNuRjt5QkFBTSxJQUFJLGVBQWUsQ0FBQyxNQUFNLEtBQUssU0FBUyxFQUFFO3dCQUM3QyxlQUFlLENBQUMsT0FBTyxDQUFDLGVBQWUsQ0FBQyxNQUFNLENBQUMsQ0FBQTtxQkFDbEQ7eUJBQU07d0JBQ0gsTUFBTSxJQUFJLEtBQUssQ0FBQyxzQkFBc0IsQ0FBQyxDQUFBO3FCQUMxQztpQkFDSjtnQkFBQyxPQUFPLEtBQUssRUFBRTtvQkFDWixJQUFJLEtBQUssQ0FBQyxPQUFPLEVBQUU7d0JBQ2YsTUFBTSxDQUFDLEtBQUssQ0FDUixxQkFBcUIsZUFBZSxDQUFDLE1BQU0sMEJBQTBCLEtBQUssQ0FBQyxPQUFPLEVBQUUsQ0FDdkYsQ0FBQTtxQkFDSjt5QkFBTTt3QkFDSCxNQUFNLENBQUMsS0FBSyxDQUFDLHFCQUFxQixlQUFlLENBQUMsTUFBTSx3QkFBd0IsQ0FBQyxDQUFBO3FCQUNwRjtpQkFDSjthQUNKO2lCQUFNO2dCQUNILE1BQU0sQ0FBQyx1QkFBdUIsQ0FBQyxlQUFlLENBQUMsQ0FBQTthQUNsRDtTQUNKO0lBQ0wsQ0FBQztJQUVELFNBQVMsa0JBQWtCLENBQUMsT0FBNEI7UUFDcEQsSUFBSSxjQUFjLEVBQUUsRUFBRTtZQUNsQixzQkFBc0I7WUFDdEIsT0FBTTtTQUNUO1FBQ0QsSUFBSSxtQkFBMkQsQ0FBQTtRQUMvRCxJQUFJLE9BQU8sQ0FBQyxNQUFNLEtBQUssa0JBQWtCLENBQUMsSUFBSSxFQUFFO1lBQzVDLG1CQUFtQixHQUFHLENBQUMsTUFBb0IsRUFBRSxFQUFFO2dCQUMzQyxNQUFNLEVBQUUsR0FBRyxNQUFNLENBQUMsRUFBRSxDQUFBO2dCQUNwQixNQUFNLE1BQU0sR0FBRyxhQUFhLENBQUMsTUFBTSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUE7Z0JBQ3hDLElBQUksTUFBTSxFQUFFO29CQUNSLE1BQU0sQ0FBQyxNQUFNLEVBQUUsQ0FBQTtpQkFDbEI7WUFDTCxDQUFDLENBQUE7U0FDSjthQUFNO1lBQ0gsTUFBTSxPQUFPLEdBQUcsb0JBQW9CLENBQUMsT0FBTyxDQUFDLE1BQU0sQ0FBQyxDQUFBO1lBQ3BELElBQUksT0FBTyxFQUFFO2dCQUNULG1CQUFtQixHQUFHLE9BQU8sQ0FBQyxPQUFPLENBQUE7YUFDeEM7U0FDSjtRQUNELElBQUksbUJBQW1CLElBQUksdUJBQXVCLEVBQUU7WUFDaEQsSUFBSTtnQkFDQSxNQUFNLENBQUMsb0JBQW9CLENBQUMsT0FBTyxDQUFDLENBQUE7Z0JBQ3BDLG1CQUFtQjtvQkFDZixDQUFDLENBQUMsbUJBQW1CLENBQUMsT0FBTyxDQUFDLE1BQU0sQ0FBQztvQkFDckMsQ0FBQyxDQUFDLHVCQUF3QixDQUFDLE9BQU8sQ0FBQyxNQUFNLEVBQUUsT0FBTyxDQUFDLE1BQU0sQ0FBQyxDQUFBO2FBQ2pFO1lBQUMsT0FBTyxLQUFLLEVBQUU7Z0JBQ1osSUFBSSxLQUFLLENBQUMsT0FBTyxFQUFFO29CQUNmLE1BQU0sQ0FBQyxLQUFLLENBQUMseUJBQXlCLE9BQU8sQ0FBQyxNQUFNLDBCQUEwQixLQUFLLENBQUMsT0FBTyxFQUFFLENBQUMsQ0FBQTtpQkFDakc7cUJBQU07b0JBQ0gsTUFBTSxDQUFDLEtBQUssQ0FBQyx5QkFBeUIsT0FBTyxDQUFDLE1BQU0sd0JBQXdCLENBQUMsQ0FBQTtpQkFDaEY7YUFDSjtTQUNKO2FBQU07WUFDSCw0QkFBNEIsQ0FBQyxJQUFJLENBQUMsT0FBTyxDQUFDLENBQUE7U0FDN0M7SUFDTCxDQUFDO0lBRUQsU0FBUyxvQkFBb0IsQ0FBQyxPQUFnQjtRQUMxQyxJQUFJLENBQUMsT0FBTyxFQUFFO1lBQ1YsTUFBTSxDQUFDLEtBQUssQ0FBQyx5QkFBeUIsQ0FBQyxDQUFBO1lBQ3ZDLE9BQU07U0FDVDtRQUNELE1BQU0sQ0FBQyxLQUFLLENBQ1IsNkVBQTZFLElBQUksQ0FBQyxTQUFTLENBQ3ZGLE9BQU8sRUFDUCxJQUFJLEVBQ0osQ0FBQyxDQUNKLEVBQUUsQ0FDTixDQUFBO1FBQ0QsbURBQW1EO1FBQ25ELE1BQU0sZUFBZSxHQUFvQixPQUEwQixDQUFBO1FBQ25FLElBQUksT0FBTyxlQUFlLENBQUMsRUFBRSxLQUFLLFFBQVEsSUFBSSxPQUFPLGVBQWUsQ0FBQyxFQUFFLEtBQUssUUFBUSxFQUFFO1lBQ2xGLE1BQU0sR0FBRyxHQUFHLE1BQU0sQ0FBQyxlQUFlLENBQUMsRUFBRSxDQUFDLENBQUE7WUFDdEMsTUFBTSxlQUFlLEdBQUcsZ0JBQWdCLENBQUMsR0FBRyxDQUFDLENBQUE7WUFDN0MsSUFBSSxlQUFlLEVBQUU7Z0JBQ2pCLGVBQWUsQ0FBQyxNQUFNLENBQUMsSUFBSSxLQUFLLENBQUMsbUVBQW1FLENBQUMsQ0FBQyxDQUFBO2FBQ3pHO1NBQ0o7SUFDTCxDQUFDO0lBRUQsU0FBUywyQkFBMkI7UUFDaEMsSUFBSSxRQUFRLEVBQUUsRUFBRTtZQUNaLE1BQU0sSUFBSSxlQUFlLENBQUMsZ0JBQWdCLENBQUMsTUFBTSxFQUFFLHVCQUF1QixDQUFDLENBQUE7U0FDOUU7UUFDRCxJQUFJLGNBQWMsRUFBRSxFQUFFO1lBQ2xCLE1BQU0sSUFBSSxlQUFlLENBQUMsZ0JBQWdCLENBQUMsWUFBWSxFQUFFLDZCQUE2QixDQUFDLENBQUE7U0FDMUY7SUFDTCxDQUFDO0lBRUQsU0FBUyxnQkFBZ0I7UUFDckIsSUFBSSxXQUFXLEVBQUUsRUFBRTtZQUNmLE1BQU0sSUFBSSxlQUFlLENBQUMsZ0JBQWdCLENBQUMsZ0JBQWdCLEVBQUUsaUNBQWlDLENBQUMsQ0FBQTtTQUNsRztJQUNMLENBQUM7SUFFRCxTQUFTLG1CQUFtQjtRQUN4QixJQUFJLENBQUMsV0FBVyxFQUFFLEVBQUU7WUFDaEIsTUFBTSxJQUFJLEtBQUssQ0FBQyxzQkFBc0IsQ0FBQyxDQUFBO1NBQzFDO0lBQ0wsQ0FBQztJQUVELE1BQU0sVUFBVSxHQUFlO1FBQzNCLGdCQUFnQixFQUFFLENBQUMsTUFBYyxFQUFFLE1BQVcsRUFBUSxFQUFFO1lBQ3BELDJCQUEyQixFQUFFLENBQUE7WUFDN0IsTUFBTSxtQkFBbUIsR0FBd0I7Z0JBQzdDLE9BQU8sRUFBRSxPQUFPO2dCQUNoQixNQUFNO2dCQUNOLE1BQU07YUFDVCxDQUFBO1lBQ0QsTUFBTSxDQUFDLGdCQUFnQixDQUFDLG1CQUFtQixDQUFDLENBQUE7WUFDNUMsVUFBVSxDQUFDLE1BQU0sQ0FBQyxLQUFLLENBQUMsbUJBQW1CLENBQUMsQ0FBQTtRQUNoRCxDQUFDO1FBQ0QsY0FBYyxFQUFFLENBQUMsSUFBc0MsRUFBRSxPQUFvQyxFQUFRLEVBQUU7WUFDbkcsMkJBQTJCLEVBQUUsQ0FBQTtZQUM3QixJQUFJLE9BQU8sSUFBSSxLQUFLLFVBQVUsRUFBRTtnQkFDNUIsdUJBQXVCLEdBQUcsSUFBSSxDQUFBO2FBQ2pDO2lCQUFNLElBQUksT0FBTyxFQUFFO2dCQUNoQixvQkFBb0IsQ0FBQyxJQUFJLENBQUMsR0FBRyxFQUFFLElBQUksRUFBRSxTQUFTLEVBQUUsT0FBTyxFQUFFLENBQUE7YUFDNUQ7UUFDTCxDQUFDO1FBQ0QsV0FBVyxFQUFFLENBQUksTUFBYyxFQUFFLE1BQVcsRUFBRSxLQUF5QixFQUFFLEVBQUU7WUFDdkUsMkJBQTJCLEVBQUUsQ0FBQTtZQUM3QixtQkFBbUIsRUFBRSxDQUFBO1lBQ3JCLEtBQUssR0FBRyxpQkFBaUIsQ0FBQyxFQUFFLENBQUMsS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLEtBQUssQ0FBQyxDQUFDLENBQUMsU0FBUyxDQUFBO1lBQ3ZELE1BQU0sRUFBRSxHQUFHLGNBQWMsRUFBRSxDQUFBO1lBQzNCLE1BQU0sTUFBTSxHQUFHLElBQUksT0FBTyxDQUFJLENBQUMsT0FBTyxFQUFFLE1BQU0sRUFBRSxFQUFFO2dCQUM5QyxNQUFNLGNBQWMsR0FBbUI7b0JBQ25DLE9BQU8sRUFBRSxPQUFPO29CQUNoQixFQUFFO29CQUNGLE1BQU07b0JBQ04sTUFBTTtpQkFDVCxDQUFBO2dCQUNELElBQUksZUFBZSxHQUEyQjtvQkFDMUMsTUFBTTtvQkFDTixPQUFPLEVBQUUsS0FBSyxLQUFLLEtBQUssQ0FBQyxPQUFPLENBQUMsQ0FBQyxDQUFDLGNBQWMsQ0FBQyxDQUFDLENBQUMsU0FBUztvQkFDN0QsVUFBVSxFQUFFLElBQUksQ0FBQyxHQUFHLEVBQUU7b0JBQ3RCLE9BQU87b0JBQ1AsTUFBTTtpQkFDVCxDQUFBO2dCQUNELE1BQU0sQ0FBQyxXQUFXLENBQUMsY0FBYyxDQUFDLENBQUE7Z0JBQ2xDLElBQUk7b0JBQ0EsVUFBVSxDQUFDLE1BQU0sQ0FBQyxLQUFLLENBQUMsY0FBYyxDQUFDLENBQUE7aUJBQzFDO2dCQUFDLE9BQU8sQ0FBQyxFQUFFO29CQUNSLGdFQUFnRTtvQkFDaEUsZUFBZSxDQUFDLE1BQU0sQ0FDbEIsSUFBSSxhQUFhLENBQU8sVUFBVSxDQUFDLGlCQUFpQixFQUFFLENBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxDQUFDLGdCQUFnQixDQUFDLENBQ2xHLENBQUE7b0JBQ0QsZUFBZSxHQUFHLElBQUksQ0FBQTtpQkFDekI7Z0JBQ0QsSUFBSSxlQUFlLEVBQUU7b0JBQ2pCLGdCQUFnQixDQUFDLE1BQU0sQ0FBQyxFQUFFLENBQUMsQ0FBQyxHQUFHLGVBQWUsQ0FBQTtpQkFDakQ7WUFDTCxDQUFDLENBQUMsQ0FBQTtZQUNGLElBQUksS0FBSyxFQUFFO2dCQUNQLEtBQUssQ0FBQyx1QkFBdUIsQ0FBQyxHQUFHLEVBQUU7b0JBQy9CLFVBQVUsQ0FBQyxnQkFBZ0IsQ0FBQyxrQkFBa0IsQ0FBQyxJQUFJLEVBQUUsRUFBRSxFQUFFLEVBQUUsQ0FBQyxDQUFBO2dCQUNoRSxDQUFDLENBQUMsQ0FBQTthQUNMO1lBQ0QsT0FBTyxNQUFNLENBQUE7UUFDakIsQ0FBQztRQUNELFNBQVMsRUFBRSxDQUFPLElBQWlDLEVBQUUsT0FBcUMsRUFBUSxFQUFFO1lBQ2hHLDJCQUEyQixFQUFFLENBQUE7WUFFN0IsSUFBSSxPQUFPLElBQUksS0FBSyxVQUFVLEVBQUU7Z0JBQzVCLGtCQUFrQixHQUFHLElBQUksQ0FBQTthQUM1QjtpQkFBTSxJQUFJLE9BQU8sRUFBRTtnQkFDaEIsZUFBZSxDQUFDLElBQUksQ0FBQyxHQUFHLEVBQUUsSUFBSSxFQUFFLFNBQVMsRUFBRSxPQUFPLEVBQUUsQ0FBQTthQUN2RDtRQUNMLENBQUM7UUFDRCxLQUFLLEVBQUUsQ0FBQyxLQUFZLEVBQUUsT0FBZSxFQUFFLGdCQUFnQixHQUFHLEtBQUssRUFBRSxFQUFFO1lBQy9ELEtBQUssR0FBRyxLQUFLLENBQUE7WUFDYixJQUFJLEtBQUssS0FBSyxLQUFLLENBQUMsR0FBRyxFQUFFO2dCQUNyQixNQUFNLEdBQUcsVUFBVSxDQUFBO2FBQ3RCO2lCQUFNO2dCQUNILE1BQU0sR0FBRyxPQUFPLENBQUE7YUFDbkI7UUFDTCxDQUFDO1FBQ0QsT0FBTyxFQUFFLFlBQVksQ0FBQyxLQUFLO1FBQzNCLE9BQU8sRUFBRSxZQUFZLENBQUMsS0FBSztRQUMzQix1QkFBdUIsRUFBRSw0QkFBNEIsQ0FBQyxLQUFLO1FBQzNELGFBQWEsRUFBRSxrQkFBa0IsQ0FBQyxLQUFLO1FBQ3ZDLFdBQVcsRUFBRSxHQUFHLEVBQUU7WUFDZCxJQUFJLGNBQWMsRUFBRSxFQUFFO2dCQUNsQixPQUFNO2FBQ1Q7WUFDRCxLQUFLLEdBQUcsZUFBZSxDQUFDLFlBQVksQ0FBQTtZQUNwQyxrQkFBa0IsQ0FBQyxJQUFJLENBQUMsU0FBUyxDQUFDLENBQUE7WUFDbEMsS0FBSyxNQUFNLEdBQUcsSUFBSSxNQUFNLENBQUMsSUFBSSxDQUFDLGdCQUFnQixDQUFDLEVBQUU7Z0JBQzdDLGdCQUFnQixDQUFDLEdBQUcsQ0FBQyxDQUFDLE1BQU0sQ0FDeEIsSUFBSSxlQUFlLENBQ2YsZ0JBQWdCLENBQUMsWUFBWSxFQUM3QixnRkFDSSxnQkFBZ0IsQ0FBQyxHQUFHLENBQUMsQ0FBQyxNQUMxQixXQUFXLENBQ2QsQ0FDSixDQUFBO2FBQ0o7WUFDRCxnQkFBZ0IsR0FBRyxNQUFNLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxDQUFBO1lBQ3RDLGFBQWEsR0FBRyxNQUFNLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxDQUFBO1lBQ25DLFlBQVksR0FBRyxJQUFJLFNBQVMsRUFBbUIsQ0FBQTtZQUMvQyxVQUFVLENBQUMsTUFBTSxDQUFDLFdBQVcsRUFBRSxDQUFBO1lBQy9CLFVBQVUsQ0FBQyxNQUFNLENBQUMsV0FBVyxFQUFFLENBQUE7UUFDbkMsQ0FBQztRQUNELE1BQU0sRUFBRSxHQUFHLEVBQUU7WUFDVCwyQkFBMkIsRUFBRSxDQUFBO1lBQzdCLGdCQUFnQixFQUFFLENBQUE7WUFFbEIsS0FBSyxHQUFHLGVBQWUsQ0FBQyxTQUFTLENBQUE7WUFDakMsVUFBVSxDQUFDLE1BQU0sQ0FBQyxNQUFNLENBQUMsUUFBUSxDQUFDLENBQUE7UUFDdEMsQ0FBQztLQUNKLENBQUE7SUFFRCxPQUFPLFVBQVUsQ0FBQTtBQUNyQixDQUFDO0FBRUQsMEVBQTBFO0FBQzFFLFNBQVMsa0JBQWtCLENBQUMsQ0FBYTtJQUNyQyxJQUFJLE9BQU8sWUFBWSxLQUFLLFdBQVcsRUFBRTtRQUNyQyxZQUFZLENBQUMsQ0FBQyxDQUFDLENBQUE7UUFDZixPQUFNO0tBQ1Q7SUFDRCxVQUFVLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFBO0FBQ3BCLENBQUMiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBVbnN1YnNjcmliYWJsZSB9IGZyb20gJ3J4anMnXG5pbXBvcnQgeyBDYW5jZWxOb3RpZmljYXRpb24sIENhbmNlbFBhcmFtcyB9IGZyb20gJy4vY2FuY2VsJ1xuaW1wb3J0IHsgQ2FuY2VsbGF0aW9uVG9rZW4sIENhbmNlbGxhdGlvblRva2VuU291cmNlIH0gZnJvbSAnLi9jYW5jZWwnXG5pbXBvcnQgeyBDb25uZWN0aW9uU3RyYXRlZ3kgfSBmcm9tICcuL2Nvbm5lY3Rpb25TdHJhdGVneSdcbmltcG9ydCB7IEVtaXR0ZXIsIEV2ZW50IH0gZnJvbSAnLi9ldmVudHMnXG5pbXBvcnQge1xuICAgIEdlbmVyaWNOb3RpZmljYXRpb25IYW5kbGVyLFxuICAgIEdlbmVyaWNSZXF1ZXN0SGFuZGxlcixcbiAgICBTdGFyTm90aWZpY2F0aW9uSGFuZGxlcixcbiAgICBTdGFyUmVxdWVzdEhhbmRsZXIsXG59IGZyb20gJy4vaGFuZGxlcnMnXG5pbXBvcnQgeyBMaW5rZWRNYXAgfSBmcm9tICcuL2xpbmtlZE1hcCdcbmltcG9ydCB7XG4gICAgRXJyb3JDb2RlcyxcbiAgICBpc05vdGlmaWNhdGlvbk1lc3NhZ2UsXG4gICAgaXNSZXF1ZXN0TWVzc2FnZSxcbiAgICBpc1Jlc3BvbnNlTWVzc2FnZSxcbiAgICBNZXNzYWdlLFxuICAgIE5vdGlmaWNhdGlvbk1lc3NhZ2UsXG4gICAgUmVxdWVzdE1lc3NhZ2UsXG4gICAgUmVzcG9uc2VFcnJvcixcbiAgICBSZXNwb25zZU1lc3NhZ2UsXG59IGZyb20gJy4vbWVzc2FnZXMnXG5pbXBvcnQgeyBub29wVHJhY2VyLCBUcmFjZSwgVHJhY2VyIH0gZnJvbSAnLi90cmFjZSdcbmltcG9ydCB7IERhdGFDYWxsYmFjaywgTWVzc2FnZVJlYWRlciwgTWVzc2FnZVdyaXRlciB9IGZyb20gJy4vdHJhbnNwb3J0J1xuXG4vLyBDb3BpZWQgZnJvbSB2c2NvZGUtbGFuZ3VhZ2VzZXJ2ZXIgdG8gYXZvaWQgYWRkaW5nIGV4dHJhbmVvdXMgZGVwZW5kZW5jaWVzLlxuXG5leHBvcnQgaW50ZXJmYWNlIExvZ2dlciB7XG4gICAgZXJyb3IobWVzc2FnZTogc3RyaW5nKTogdm9pZFxuICAgIHdhcm4obWVzc2FnZTogc3RyaW5nKTogdm9pZFxuICAgIGluZm8obWVzc2FnZTogc3RyaW5nKTogdm9pZFxuICAgIGxvZyhtZXNzYWdlOiBzdHJpbmcpOiB2b2lkXG59XG5cbmNvbnN0IE51bGxMb2dnZXI6IExvZ2dlciA9IE9iamVjdC5mcmVlemUoe1xuICAgIGVycm9yOiAoKSA9PiB7XG4gICAgICAgIC8qIG5vb3AgKi9cbiAgICB9LFxuICAgIHdhcm46ICgpID0+IHtcbiAgICAgICAgLyogbm9vcCAqL1xuICAgIH0sXG4gICAgaW5mbzogKCkgPT4ge1xuICAgICAgICAvKiBub29wICovXG4gICAgfSxcbiAgICBsb2c6ICgpID0+IHtcbiAgICAgICAgLyogbm9vcCAqL1xuICAgIH0sXG59KVxuXG5leHBvcnQgZW51bSBDb25uZWN0aW9uRXJyb3JzIHtcbiAgICAvKipcbiAgICAgKiBUaGUgY29ubmVjdGlvbiBpcyBjbG9zZWQuXG4gICAgICovXG4gICAgQ2xvc2VkID0gMSxcbiAgICAvKipcbiAgICAgKiBUaGUgY29ubmVjdGlvbiBnb3QgdW5zdWJzY3JpYmVkIChpLmUuLCBkaXNwb3NlZCkuXG4gICAgICovXG4gICAgVW5zdWJzY3JpYmVkID0gMixcbiAgICAvKipcbiAgICAgKiBUaGUgY29ubmVjdGlvbiBpcyBhbHJlYWR5IGluIGxpc3RlbmluZyBtb2RlLlxuICAgICAqL1xuICAgIEFscmVhZHlMaXN0ZW5pbmcgPSAzLFxufVxuXG5leHBvcnQgY2xhc3MgQ29ubmVjdGlvbkVycm9yIGV4dGVuZHMgRXJyb3Ige1xuICAgIHB1YmxpYyByZWFkb25seSBjb2RlOiBDb25uZWN0aW9uRXJyb3JzXG5cbiAgICBjb25zdHJ1Y3Rvcihjb2RlOiBDb25uZWN0aW9uRXJyb3JzLCBtZXNzYWdlOiBzdHJpbmcpIHtcbiAgICAgICAgc3VwZXIobWVzc2FnZSlcbiAgICAgICAgdGhpcy5jb2RlID0gY29kZVxuICAgICAgICBPYmplY3Quc2V0UHJvdG90eXBlT2YodGhpcywgQ29ubmVjdGlvbkVycm9yLnByb3RvdHlwZSlcbiAgICB9XG59XG5cbnR5cGUgTWVzc2FnZVF1ZXVlID0gTGlua2VkTWFwPHN0cmluZywgTWVzc2FnZT5cblxuZXhwb3J0IGludGVyZmFjZSBDb25uZWN0aW9uIGV4dGVuZHMgVW5zdWJzY3JpYmFibGUge1xuICAgIHNlbmRSZXF1ZXN0PFI+KG1ldGhvZDogc3RyaW5nLCBwYXJhbXM/OiBhbnkpOiBQcm9taXNlPFI+XG4gICAgc2VuZFJlcXVlc3Q8Uj4obWV0aG9kOiBzdHJpbmcsIC4uLnBhcmFtczogYW55W10pOiBQcm9taXNlPFI+XG5cbiAgICBvblJlcXVlc3Q8UiwgRT4obWV0aG9kOiBzdHJpbmcsIGhhbmRsZXI6IEdlbmVyaWNSZXF1ZXN0SGFuZGxlcjxSLCBFPik6IHZvaWRcbiAgICBvblJlcXVlc3QoaGFuZGxlcjogU3RhclJlcXVlc3RIYW5kbGVyKTogdm9pZFxuXG4gICAgc2VuZE5vdGlmaWNhdGlvbihtZXRob2Q6IHN0cmluZywgLi4ucGFyYW1zOiBhbnlbXSk6IHZvaWRcblxuICAgIG9uTm90aWZpY2F0aW9uKG1ldGhvZDogc3RyaW5nLCBoYW5kbGVyOiBHZW5lcmljTm90aWZpY2F0aW9uSGFuZGxlcik6IHZvaWRcbiAgICBvbk5vdGlmaWNhdGlvbihoYW5kbGVyOiBTdGFyTm90aWZpY2F0aW9uSGFuZGxlcik6IHZvaWRcblxuICAgIHRyYWNlKHZhbHVlOiBUcmFjZSwgdHJhY2VyOiBUcmFjZXIpOiB2b2lkXG5cbiAgICBvbkVycm9yOiBFdmVudDxbRXJyb3IsIE1lc3NhZ2UgfCB1bmRlZmluZWQsIG51bWJlciB8IHVuZGVmaW5lZF0+XG4gICAgb25DbG9zZTogRXZlbnQ8dm9pZD5cbiAgICBvblVuaGFuZGxlZE5vdGlmaWNhdGlvbjogRXZlbnQ8Tm90aWZpY2F0aW9uTWVzc2FnZT5cbiAgICBsaXN0ZW4oKTogdm9pZFxuICAgIG9uVW5zdWJzY3JpYmU6IEV2ZW50PHZvaWQ+XG59XG5cbmV4cG9ydCBpbnRlcmZhY2UgTWVzc2FnZVRyYW5zcG9ydHMge1xuICAgIHJlYWRlcjogTWVzc2FnZVJlYWRlclxuICAgIHdyaXRlcjogTWVzc2FnZVdyaXRlclxufVxuXG5leHBvcnQgZnVuY3Rpb24gY3JlYXRlQ29ubmVjdGlvbihcbiAgICB0cmFuc3BvcnRzOiBNZXNzYWdlVHJhbnNwb3J0cyxcbiAgICBsb2dnZXI/OiBMb2dnZXIsXG4gICAgc3RyYXRlZ3k/OiBDb25uZWN0aW9uU3RyYXRlZ3lcbik6IENvbm5lY3Rpb24ge1xuICAgIGlmICghbG9nZ2VyKSB7XG4gICAgICAgIGxvZ2dlciA9IE51bGxMb2dnZXJcbiAgICB9XG4gICAgcmV0dXJuIF9jcmVhdGVDb25uZWN0aW9uKHRyYW5zcG9ydHMsIGxvZ2dlciwgc3RyYXRlZ3kpXG59XG5cbmludGVyZmFjZSBSZXNwb25zZVByb21pc2Uge1xuICAgIC8qKiBUaGUgcmVxdWVzdCdzIG1ldGhvZC4gKi9cbiAgICBtZXRob2Q6IHN0cmluZ1xuXG4gICAgLyoqIE9ubHkgc2V0IGluIFRyYWNlLlZlcmJvc2UgbW9kZS4gKi9cbiAgICByZXF1ZXN0PzogUmVxdWVzdE1lc3NhZ2VcblxuICAgIC8qKiBUaGUgdGltZXN0YW1wIHdoZW4gdGhlIHJlcXVlc3Qgd2FzIHJlY2VpdmVkLiAqL1xuICAgIHRpbWVyU3RhcnQ6IG51bWJlclxuXG4gICAgcmVzb2x2ZTogKHJlc3BvbnNlOiBhbnkpID0+IHZvaWRcbiAgICByZWplY3Q6IChlcnJvcjogYW55KSA9PiB2b2lkXG59XG5cbmVudW0gQ29ubmVjdGlvblN0YXRlIHtcbiAgICBOZXcgPSAxLFxuICAgIExpc3RlbmluZyA9IDIsXG4gICAgQ2xvc2VkID0gMyxcbiAgICBVbnN1YnNjcmliZWQgPSA0LFxufVxuXG5pbnRlcmZhY2UgUmVxdWVzdEhhbmRsZXJFbGVtZW50IHtcbiAgICB0eXBlOiBzdHJpbmcgfCB1bmRlZmluZWRcbiAgICBoYW5kbGVyOiBHZW5lcmljUmVxdWVzdEhhbmRsZXI8YW55LCBhbnk+XG59XG5cbmludGVyZmFjZSBOb3RpZmljYXRpb25IYW5kbGVyRWxlbWVudCB7XG4gICAgdHlwZTogc3RyaW5nIHwgdW5kZWZpbmVkXG4gICAgaGFuZGxlcjogR2VuZXJpY05vdGlmaWNhdGlvbkhhbmRsZXJcbn1cblxuZnVuY3Rpb24gX2NyZWF0ZUNvbm5lY3Rpb24odHJhbnNwb3J0czogTWVzc2FnZVRyYW5zcG9ydHMsIGxvZ2dlcjogTG9nZ2VyLCBzdHJhdGVneT86IENvbm5lY3Rpb25TdHJhdGVneSk6IENvbm5lY3Rpb24ge1xuICAgIGxldCBzZXF1ZW5jZU51bWJlciA9IDBcbiAgICBsZXQgbm90aWZpY2F0aW9uU3F1ZW5jZU51bWJlciA9IDBcbiAgICBsZXQgdW5rbm93blJlc3BvbnNlU3F1ZW5jZU51bWJlciA9IDBcbiAgICBjb25zdCB2ZXJzaW9uID0gJzIuMCdcblxuICAgIGxldCBzdGFyUmVxdWVzdEhhbmRsZXI6IFN0YXJSZXF1ZXN0SGFuZGxlciB8IHVuZGVmaW5lZFxuICAgIGNvbnN0IHJlcXVlc3RIYW5kbGVyczogeyBbbmFtZTogc3RyaW5nXTogUmVxdWVzdEhhbmRsZXJFbGVtZW50IHwgdW5kZWZpbmVkIH0gPSBPYmplY3QuY3JlYXRlKG51bGwpXG4gICAgbGV0IHN0YXJOb3RpZmljYXRpb25IYW5kbGVyOiBTdGFyTm90aWZpY2F0aW9uSGFuZGxlciB8IHVuZGVmaW5lZFxuICAgIGNvbnN0IG5vdGlmaWNhdGlvbkhhbmRsZXJzOiB7IFtuYW1lOiBzdHJpbmddOiBOb3RpZmljYXRpb25IYW5kbGVyRWxlbWVudCB8IHVuZGVmaW5lZCB9ID0gT2JqZWN0LmNyZWF0ZShudWxsKVxuXG4gICAgbGV0IHRpbWVyID0gZmFsc2VcbiAgICBsZXQgbWVzc2FnZVF1ZXVlOiBNZXNzYWdlUXVldWUgPSBuZXcgTGlua2VkTWFwPHN0cmluZywgTWVzc2FnZT4oKVxuICAgIGxldCByZXNwb25zZVByb21pc2VzOiB7IFtuYW1lOiBzdHJpbmddOiBSZXNwb25zZVByb21pc2UgfSA9IE9iamVjdC5jcmVhdGUobnVsbClcbiAgICBsZXQgcmVxdWVzdFRva2VuczogeyBbaWQ6IHN0cmluZ106IENhbmNlbGxhdGlvblRva2VuU291cmNlIH0gPSBPYmplY3QuY3JlYXRlKG51bGwpXG5cbiAgICBsZXQgdHJhY2U6IFRyYWNlID0gVHJhY2UuT2ZmXG4gICAgbGV0IHRyYWNlcjogVHJhY2VyID0gbm9vcFRyYWNlclxuXG4gICAgbGV0IHN0YXRlOiBDb25uZWN0aW9uU3RhdGUgPSBDb25uZWN0aW9uU3RhdGUuTmV3XG4gICAgY29uc3QgZXJyb3JFbWl0dGVyID0gbmV3IEVtaXR0ZXI8W0Vycm9yLCBNZXNzYWdlIHwgdW5kZWZpbmVkLCBudW1iZXIgfCB1bmRlZmluZWRdPigpXG4gICAgY29uc3QgY2xvc2VFbWl0dGVyOiBFbWl0dGVyPHZvaWQ+ID0gbmV3IEVtaXR0ZXI8dm9pZD4oKVxuICAgIGNvbnN0IHVuaGFuZGxlZE5vdGlmaWNhdGlvbkVtaXR0ZXI6IEVtaXR0ZXI8Tm90aWZpY2F0aW9uTWVzc2FnZT4gPSBuZXcgRW1pdHRlcjxOb3RpZmljYXRpb25NZXNzYWdlPigpXG5cbiAgICBjb25zdCB1bnN1YnNjcmliZUVtaXR0ZXI6IEVtaXR0ZXI8dm9pZD4gPSBuZXcgRW1pdHRlcjx2b2lkPigpXG5cbiAgICBmdW5jdGlvbiBjcmVhdGVSZXF1ZXN0UXVldWVLZXkoaWQ6IHN0cmluZyB8IG51bWJlcik6IHN0cmluZyB7XG4gICAgICAgIHJldHVybiAncmVxLScgKyBpZC50b1N0cmluZygpXG4gICAgfVxuXG4gICAgZnVuY3Rpb24gY3JlYXRlUmVzcG9uc2VRdWV1ZUtleShpZDogc3RyaW5nIHwgbnVtYmVyIHwgbnVsbCk6IHN0cmluZyB7XG4gICAgICAgIGlmIChpZCA9PT0gbnVsbCkge1xuICAgICAgICAgICAgcmV0dXJuICdyZXMtdW5rbm93bi0nICsgKCsrdW5rbm93blJlc3BvbnNlU3F1ZW5jZU51bWJlcikudG9TdHJpbmcoKVxuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgcmV0dXJuICdyZXMtJyArIGlkLnRvU3RyaW5nKClcbiAgICAgICAgfVxuICAgIH1cblxuICAgIGZ1bmN0aW9uIGNyZWF0ZU5vdGlmaWNhdGlvblF1ZXVlS2V5KCk6IHN0cmluZyB7XG4gICAgICAgIHJldHVybiAnbm90LScgKyAoKytub3RpZmljYXRpb25TcXVlbmNlTnVtYmVyKS50b1N0cmluZygpXG4gICAgfVxuXG4gICAgZnVuY3Rpb24gYWRkTWVzc2FnZVRvUXVldWUocXVldWU6IE1lc3NhZ2VRdWV1ZSwgbWVzc2FnZTogTWVzc2FnZSk6IHZvaWQge1xuICAgICAgICBpZiAoaXNSZXF1ZXN0TWVzc2FnZShtZXNzYWdlKSkge1xuICAgICAgICAgICAgcXVldWUuc2V0KGNyZWF0ZVJlcXVlc3RRdWV1ZUtleShtZXNzYWdlLmlkKSwgbWVzc2FnZSlcbiAgICAgICAgfSBlbHNlIGlmIChpc1Jlc3BvbnNlTWVzc2FnZShtZXNzYWdlKSkge1xuICAgICAgICAgICAgcXVldWUuc2V0KGNyZWF0ZVJlc3BvbnNlUXVldWVLZXkobWVzc2FnZS5pZCksIG1lc3NhZ2UpXG4gICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICBxdWV1ZS5zZXQoY3JlYXRlTm90aWZpY2F0aW9uUXVldWVLZXkoKSwgbWVzc2FnZSlcbiAgICAgICAgfVxuICAgIH1cblxuICAgIGZ1bmN0aW9uIGNhbmNlbFVuZGlzcGF0Y2hlZChfbWVzc2FnZTogTWVzc2FnZSk6IFJlc3BvbnNlTWVzc2FnZSB8IHVuZGVmaW5lZCB7XG4gICAgICAgIHJldHVybiB1bmRlZmluZWRcbiAgICB9XG5cbiAgICBmdW5jdGlvbiBpc0xpc3RlbmluZygpOiBib29sZWFuIHtcbiAgICAgICAgcmV0dXJuIHN0YXRlID09PSBDb25uZWN0aW9uU3RhdGUuTGlzdGVuaW5nXG4gICAgfVxuXG4gICAgZnVuY3Rpb24gaXNDbG9zZWQoKTogYm9vbGVhbiB7XG4gICAgICAgIHJldHVybiBzdGF0ZSA9PT0gQ29ubmVjdGlvblN0YXRlLkNsb3NlZFxuICAgIH1cblxuICAgIGZ1bmN0aW9uIGlzVW5zdWJzY3JpYmVkKCk6IGJvb2xlYW4ge1xuICAgICAgICByZXR1cm4gc3RhdGUgPT09IENvbm5lY3Rpb25TdGF0ZS5VbnN1YnNjcmliZWRcbiAgICB9XG5cbiAgICBmdW5jdGlvbiBjbG9zZUhhbmRsZXIoKTogdm9pZCB7XG4gICAgICAgIGlmIChzdGF0ZSA9PT0gQ29ubmVjdGlvblN0YXRlLk5ldyB8fCBzdGF0ZSA9PT0gQ29ubmVjdGlvblN0YXRlLkxpc3RlbmluZykge1xuICAgICAgICAgICAgc3RhdGUgPSBDb25uZWN0aW9uU3RhdGUuQ2xvc2VkXG4gICAgICAgICAgICBjbG9zZUVtaXR0ZXIuZmlyZSh1bmRlZmluZWQpXG4gICAgICAgIH1cbiAgICAgICAgLy8gSWYgdGhlIGNvbm5lY3Rpb24gaXMgdW5zdWJzY3JpYmVkIGRvbid0IHNlbnQgY2xvc2UgZXZlbnRzLlxuICAgIH1cblxuICAgIGZ1bmN0aW9uIHJlYWRFcnJvckhhbmRsZXIoZXJyb3I6IEVycm9yKTogdm9pZCB7XG4gICAgICAgIGVycm9yRW1pdHRlci5maXJlKFtlcnJvciwgdW5kZWZpbmVkLCB1bmRlZmluZWRdKVxuICAgIH1cblxuICAgIGZ1bmN0aW9uIHdyaXRlRXJyb3JIYW5kbGVyKGRhdGE6IFtFcnJvciwgTWVzc2FnZSB8IHVuZGVmaW5lZCwgbnVtYmVyIHwgdW5kZWZpbmVkXSk6IHZvaWQge1xuICAgICAgICBlcnJvckVtaXR0ZXIuZmlyZShkYXRhKVxuICAgIH1cblxuICAgIHRyYW5zcG9ydHMucmVhZGVyLm9uQ2xvc2UoY2xvc2VIYW5kbGVyKVxuICAgIHRyYW5zcG9ydHMucmVhZGVyLm9uRXJyb3IocmVhZEVycm9ySGFuZGxlcilcblxuICAgIHRyYW5zcG9ydHMud3JpdGVyLm9uQ2xvc2UoY2xvc2VIYW5kbGVyKVxuICAgIHRyYW5zcG9ydHMud3JpdGVyLm9uRXJyb3Iod3JpdGVFcnJvckhhbmRsZXIpXG5cbiAgICBmdW5jdGlvbiB0cmlnZ2VyTWVzc2FnZVF1ZXVlKCk6IHZvaWQge1xuICAgICAgICBpZiAodGltZXIgfHwgbWVzc2FnZVF1ZXVlLnNpemUgPT09IDApIHtcbiAgICAgICAgICAgIHJldHVyblxuICAgICAgICB9XG4gICAgICAgIHRpbWVyID0gdHJ1ZVxuICAgICAgICBzZXRJbW1lZGlhdGVDb21wYXQoKCkgPT4ge1xuICAgICAgICAgICAgdGltZXIgPSBmYWxzZVxuICAgICAgICAgICAgcHJvY2Vzc01lc3NhZ2VRdWV1ZSgpXG4gICAgICAgIH0pXG4gICAgfVxuXG4gICAgZnVuY3Rpb24gcHJvY2Vzc01lc3NhZ2VRdWV1ZSgpOiB2b2lkIHtcbiAgICAgICAgaWYgKG1lc3NhZ2VRdWV1ZS5zaXplID09PSAwKSB7XG4gICAgICAgICAgICByZXR1cm5cbiAgICAgICAgfVxuICAgICAgICBjb25zdCBtZXNzYWdlID0gbWVzc2FnZVF1ZXVlLnNoaWZ0KCkhXG4gICAgICAgIHRyeSB7XG4gICAgICAgICAgICBpZiAoaXNSZXF1ZXN0TWVzc2FnZShtZXNzYWdlKSkge1xuICAgICAgICAgICAgICAgIGhhbmRsZVJlcXVlc3QobWVzc2FnZSlcbiAgICAgICAgICAgIH0gZWxzZSBpZiAoaXNOb3RpZmljYXRpb25NZXNzYWdlKG1lc3NhZ2UpKSB7XG4gICAgICAgICAgICAgICAgaGFuZGxlTm90aWZpY2F0aW9uKG1lc3NhZ2UpXG4gICAgICAgICAgICB9IGVsc2UgaWYgKGlzUmVzcG9uc2VNZXNzYWdlKG1lc3NhZ2UpKSB7XG4gICAgICAgICAgICAgICAgaGFuZGxlUmVzcG9uc2UobWVzc2FnZSlcbiAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgaGFuZGxlSW52YWxpZE1lc3NhZ2UobWVzc2FnZSlcbiAgICAgICAgICAgIH1cbiAgICAgICAgfSBmaW5hbGx5IHtcbiAgICAgICAgICAgIHRyaWdnZXJNZXNzYWdlUXVldWUoKVxuICAgICAgICB9XG4gICAgfVxuXG4gICAgY29uc3QgY2FsbGJhY2s6IERhdGFDYWxsYmFjayA9IG1lc3NhZ2UgPT4ge1xuICAgICAgICB0cnkge1xuICAgICAgICAgICAgLy8gV2UgaGF2ZSByZWNlaXZlZCBhIGNhbmNlbGxhdGlvbiBtZXNzYWdlLiBDaGVjayBpZiB0aGUgbWVzc2FnZSBpcyBzdGlsbCBpbiB0aGUgcXVldWUgYW5kIGNhbmNlbCBpdCBpZlxuICAgICAgICAgICAgLy8gYWxsb3dlZCB0byBkbyBzby5cbiAgICAgICAgICAgIGlmIChpc05vdGlmaWNhdGlvbk1lc3NhZ2UobWVzc2FnZSkgJiYgbWVzc2FnZS5tZXRob2QgPT09IENhbmNlbE5vdGlmaWNhdGlvbi50eXBlKSB7XG4gICAgICAgICAgICAgICAgY29uc3Qga2V5ID0gY3JlYXRlUmVxdWVzdFF1ZXVlS2V5KChtZXNzYWdlLnBhcmFtcyBhcyBDYW5jZWxQYXJhbXMpLmlkKVxuICAgICAgICAgICAgICAgIGNvbnN0IHRvQ2FuY2VsID0gbWVzc2FnZVF1ZXVlLmdldChrZXkpXG4gICAgICAgICAgICAgICAgaWYgKGlzUmVxdWVzdE1lc3NhZ2UodG9DYW5jZWwpKSB7XG4gICAgICAgICAgICAgICAgICAgIGNvbnN0IHJlc3BvbnNlID1cbiAgICAgICAgICAgICAgICAgICAgICAgIHN0cmF0ZWd5ICYmIHN0cmF0ZWd5LmNhbmNlbFVuZGlzcGF0Y2hlZFxuICAgICAgICAgICAgICAgICAgICAgICAgICAgID8gc3RyYXRlZ3kuY2FuY2VsVW5kaXNwYXRjaGVkKHRvQ2FuY2VsLCBjYW5jZWxVbmRpc3BhdGNoZWQpXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgOiBjYW5jZWxVbmRpc3BhdGNoZWQodG9DYW5jZWwpXG4gICAgICAgICAgICAgICAgICAgIGlmIChyZXNwb25zZSAmJiAocmVzcG9uc2UuZXJyb3IgIT09IHVuZGVmaW5lZCB8fCByZXNwb25zZS5yZXN1bHQgIT09IHVuZGVmaW5lZCkpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgIG1lc3NhZ2VRdWV1ZS5kZWxldGUoa2V5KVxuICAgICAgICAgICAgICAgICAgICAgICAgcmVzcG9uc2UuaWQgPSB0b0NhbmNlbC5pZFxuICAgICAgICAgICAgICAgICAgICAgICAgdHJhY2VyLnJlc3BvbnNlQ2FuY2VsZWQocmVzcG9uc2UsIHRvQ2FuY2VsLCBtZXNzYWdlKVxuICAgICAgICAgICAgICAgICAgICAgICAgdHJhbnNwb3J0cy53cml0ZXIud3JpdGUocmVzcG9uc2UpXG4gICAgICAgICAgICAgICAgICAgICAgICByZXR1cm5cbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIGFkZE1lc3NhZ2VUb1F1ZXVlKG1lc3NhZ2VRdWV1ZSwgbWVzc2FnZSlcbiAgICAgICAgfSBmaW5hbGx5IHtcbiAgICAgICAgICAgIHRyaWdnZXJNZXNzYWdlUXVldWUoKVxuICAgICAgICB9XG4gICAgfVxuXG4gICAgZnVuY3Rpb24gaGFuZGxlUmVxdWVzdChyZXF1ZXN0TWVzc2FnZTogUmVxdWVzdE1lc3NhZ2UpOiB2b2lkIHtcbiAgICAgICAgaWYgKGlzVW5zdWJzY3JpYmVkKCkpIHtcbiAgICAgICAgICAgIC8vIHdlIHJldHVybiBoZXJlIHNpbGVudGx5IHNpbmNlIHdlIGZpcmVkIGFuIGV2ZW50IHdoZW4gdGhlXG4gICAgICAgICAgICAvLyBjb25uZWN0aW9uIGdvdCB1bnN1YnNjcmliZWQuXG4gICAgICAgICAgICByZXR1cm5cbiAgICAgICAgfVxuXG4gICAgICAgIGNvbnN0IHN0YXJ0VGltZSA9IERhdGUubm93KClcblxuICAgICAgICBmdW5jdGlvbiByZXBseShyZXN1bHRPckVycm9yOiBhbnkgfCBSZXNwb25zZUVycm9yPGFueT4pOiB2b2lkIHtcbiAgICAgICAgICAgIGNvbnN0IG1lc3NhZ2U6IFJlc3BvbnNlTWVzc2FnZSA9IHtcbiAgICAgICAgICAgICAgICBqc29ucnBjOiB2ZXJzaW9uLFxuICAgICAgICAgICAgICAgIGlkOiByZXF1ZXN0TWVzc2FnZS5pZCxcbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIGlmIChyZXN1bHRPckVycm9yIGluc3RhbmNlb2YgUmVzcG9uc2VFcnJvcikge1xuICAgICAgICAgICAgICAgIG1lc3NhZ2UuZXJyb3IgPSAocmVzdWx0T3JFcnJvciBhcyBSZXNwb25zZUVycm9yPGFueT4pLnRvSlNPTigpXG4gICAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgICAgIG1lc3NhZ2UucmVzdWx0ID0gcmVzdWx0T3JFcnJvciA9PT0gdW5kZWZpbmVkID8gbnVsbCA6IHJlc3VsdE9yRXJyb3JcbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIHRyYWNlci5yZXNwb25zZVNlbnQobWVzc2FnZSwgcmVxdWVzdE1lc3NhZ2UsIHN0YXJ0VGltZSlcbiAgICAgICAgICAgIHRyYW5zcG9ydHMud3JpdGVyLndyaXRlKG1lc3NhZ2UpXG4gICAgICAgIH1cbiAgICAgICAgZnVuY3Rpb24gcmVwbHlFcnJvcihlcnJvcjogUmVzcG9uc2VFcnJvcjxhbnk+KTogdm9pZCB7XG4gICAgICAgICAgICBjb25zdCBtZXNzYWdlOiBSZXNwb25zZU1lc3NhZ2UgPSB7XG4gICAgICAgICAgICAgICAganNvbnJwYzogdmVyc2lvbixcbiAgICAgICAgICAgICAgICBpZDogcmVxdWVzdE1lc3NhZ2UuaWQsXG4gICAgICAgICAgICAgICAgZXJyb3I6IGVycm9yLnRvSlNPTigpLFxuICAgICAgICAgICAgfVxuICAgICAgICAgICAgdHJhY2VyLnJlc3BvbnNlU2VudChtZXNzYWdlLCByZXF1ZXN0TWVzc2FnZSwgc3RhcnRUaW1lKVxuICAgICAgICAgICAgdHJhbnNwb3J0cy53cml0ZXIud3JpdGUobWVzc2FnZSlcbiAgICAgICAgfVxuICAgICAgICBmdW5jdGlvbiByZXBseVN1Y2Nlc3MocmVzdWx0OiBhbnkpOiB2b2lkIHtcbiAgICAgICAgICAgIC8vIFRoZSBKU09OIFJQQyBkZWZpbmVzIHRoYXQgYSByZXNwb25zZSBtdXN0IGVpdGhlciBoYXZlIGEgcmVzdWx0IG9yIGFuIGVycm9yXG4gICAgICAgICAgICAvLyBTbyB3ZSBjYW4ndCB0cmVhdCB1bmRlZmluZWQgYXMgYSB2YWxpZCByZXNwb25zZSByZXN1bHQuXG4gICAgICAgICAgICBpZiAocmVzdWx0ID09PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICAgICAgICByZXN1bHQgPSBudWxsXG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBjb25zdCBtZXNzYWdlOiBSZXNwb25zZU1lc3NhZ2UgPSB7XG4gICAgICAgICAgICAgICAganNvbnJwYzogdmVyc2lvbixcbiAgICAgICAgICAgICAgICBpZDogcmVxdWVzdE1lc3NhZ2UuaWQsXG4gICAgICAgICAgICAgICAgcmVzdWx0LFxuICAgICAgICAgICAgfVxuICAgICAgICAgICAgdHJhY2VyLnJlc3BvbnNlU2VudChtZXNzYWdlLCByZXF1ZXN0TWVzc2FnZSwgc3RhcnRUaW1lKVxuICAgICAgICAgICAgdHJhbnNwb3J0cy53cml0ZXIud3JpdGUobWVzc2FnZSlcbiAgICAgICAgfVxuXG4gICAgICAgIHRyYWNlci5yZXF1ZXN0UmVjZWl2ZWQocmVxdWVzdE1lc3NhZ2UpXG5cbiAgICAgICAgY29uc3QgZWxlbWVudCA9IHJlcXVlc3RIYW5kbGVyc1tyZXF1ZXN0TWVzc2FnZS5tZXRob2RdXG4gICAgICAgIGNvbnN0IHJlcXVlc3RIYW5kbGVyOiBHZW5lcmljUmVxdWVzdEhhbmRsZXI8YW55LCBhbnk+IHwgdW5kZWZpbmVkID0gZWxlbWVudCAmJiBlbGVtZW50LmhhbmRsZXJcbiAgICAgICAgaWYgKHJlcXVlc3RIYW5kbGVyIHx8IHN0YXJSZXF1ZXN0SGFuZGxlcikge1xuICAgICAgICAgICAgY29uc3QgY2FuY2VsbGF0aW9uU291cmNlID0gbmV3IENhbmNlbGxhdGlvblRva2VuU291cmNlKClcbiAgICAgICAgICAgIGNvbnN0IHRva2VuS2V5ID0gU3RyaW5nKHJlcXVlc3RNZXNzYWdlLmlkKVxuICAgICAgICAgICAgcmVxdWVzdFRva2Vuc1t0b2tlbktleV0gPSBjYW5jZWxsYXRpb25Tb3VyY2VcbiAgICAgICAgICAgIHRyeSB7XG4gICAgICAgICAgICAgICAgY29uc3QgcGFyYW1zID0gcmVxdWVzdE1lc3NhZ2UucGFyYW1zICE9PSB1bmRlZmluZWQgPyByZXF1ZXN0TWVzc2FnZS5wYXJhbXMgOiBudWxsXG4gICAgICAgICAgICAgICAgY29uc3QgaGFuZGxlclJlc3VsdCA9IHJlcXVlc3RIYW5kbGVyXG4gICAgICAgICAgICAgICAgICAgID8gcmVxdWVzdEhhbmRsZXIocGFyYW1zLCBjYW5jZWxsYXRpb25Tb3VyY2UudG9rZW4pXG4gICAgICAgICAgICAgICAgICAgIDogc3RhclJlcXVlc3RIYW5kbGVyIShyZXF1ZXN0TWVzc2FnZS5tZXRob2QsIHBhcmFtcywgY2FuY2VsbGF0aW9uU291cmNlLnRva2VuKVxuXG4gICAgICAgICAgICAgICAgY29uc3QgcHJvbWlzZSA9IGhhbmRsZXJSZXN1bHQgYXMgUHJvbWlzZTxhbnkgfCBSZXNwb25zZUVycm9yPGFueT4+XG4gICAgICAgICAgICAgICAgaWYgKCFoYW5kbGVyUmVzdWx0KSB7XG4gICAgICAgICAgICAgICAgICAgIGRlbGV0ZSByZXF1ZXN0VG9rZW5zW3Rva2VuS2V5XVxuICAgICAgICAgICAgICAgICAgICByZXBseVN1Y2Nlc3MoaGFuZGxlclJlc3VsdClcbiAgICAgICAgICAgICAgICB9IGVsc2UgaWYgKHByb21pc2UudGhlbikge1xuICAgICAgICAgICAgICAgICAgICBwcm9taXNlLnRoZW4oXG4gICAgICAgICAgICAgICAgICAgICAgICAocmVzdWx0T3JFcnJvcik6IGFueSB8IFJlc3BvbnNlRXJyb3I8YW55PiA9PiB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgZGVsZXRlIHJlcXVlc3RUb2tlbnNbdG9rZW5LZXldXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgcmVwbHkocmVzdWx0T3JFcnJvcilcbiAgICAgICAgICAgICAgICAgICAgICAgIH0sXG4gICAgICAgICAgICAgICAgICAgICAgICBlcnJvciA9PiB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgZGVsZXRlIHJlcXVlc3RUb2tlbnNbdG9rZW5LZXldXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgaWYgKGVycm9yIGluc3RhbmNlb2YgUmVzcG9uc2VFcnJvcikge1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICByZXBseUVycm9yKGVycm9yIGFzIFJlc3BvbnNlRXJyb3I8YW55PilcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICB9IGVsc2UgaWYgKGVycm9yICYmIHR5cGVvZiBlcnJvci5tZXNzYWdlID09PSAnc3RyaW5nJykge1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICByZXBseUVycm9yKFxuICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgbmV3IFJlc3BvbnNlRXJyb3I8dm9pZD4oRXJyb3JDb2Rlcy5JbnRlcm5hbEVycm9yLCBlcnJvci5tZXNzYWdlLCB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgc3RhY2s6IGVycm9yLnN0YWNrLFxuICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIC4uLmVycm9yLFxuICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgfSlcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIHJlcGx5RXJyb3IoXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICBuZXcgUmVzcG9uc2VFcnJvcjx2b2lkPihcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICBFcnJvckNvZGVzLkludGVybmFsRXJyb3IsXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgYFJlcXVlc3QgJHtcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgcmVxdWVzdE1lc3NhZ2UubWV0aG9kXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgfSBmYWlsZWQgdW5leHBlY3RlZGx5IHdpdGhvdXQgcHJvdmlkaW5nIGFueSBkZXRhaWxzLmBcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgIGRlbGV0ZSByZXF1ZXN0VG9rZW5zW3Rva2VuS2V5XVxuICAgICAgICAgICAgICAgICAgICByZXBseShoYW5kbGVyUmVzdWx0KVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH0gY2F0Y2ggKGVycm9yKSB7XG4gICAgICAgICAgICAgICAgZGVsZXRlIHJlcXVlc3RUb2tlbnNbdG9rZW5LZXldXG4gICAgICAgICAgICAgICAgaWYgKGVycm9yIGluc3RhbmNlb2YgUmVzcG9uc2VFcnJvcikge1xuICAgICAgICAgICAgICAgICAgICByZXBseShlcnJvciBhcyBSZXNwb25zZUVycm9yPGFueT4pXG4gICAgICAgICAgICAgICAgfSBlbHNlIGlmIChlcnJvciAmJiB0eXBlb2YgZXJyb3IubWVzc2FnZSA9PT0gJ3N0cmluZycpIHtcbiAgICAgICAgICAgICAgICAgICAgcmVwbHlFcnJvcihcbiAgICAgICAgICAgICAgICAgICAgICAgIG5ldyBSZXNwb25zZUVycm9yPHZvaWQ+KEVycm9yQ29kZXMuSW50ZXJuYWxFcnJvciwgZXJyb3IubWVzc2FnZSwge1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgIHN0YWNrOiBlcnJvci5zdGFjayxcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAuLi5lcnJvcixcbiAgICAgICAgICAgICAgICAgICAgICAgIH0pXG4gICAgICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgICAgICAgICByZXBseUVycm9yKFxuICAgICAgICAgICAgICAgICAgICAgICAgbmV3IFJlc3BvbnNlRXJyb3I8dm9pZD4oXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgRXJyb3JDb2Rlcy5JbnRlcm5hbEVycm9yLFxuICAgICAgICAgICAgICAgICAgICAgICAgICAgIGBSZXF1ZXN0ICR7cmVxdWVzdE1lc3NhZ2UubWV0aG9kfSBmYWlsZWQgdW5leHBlY3RlZGx5IHdpdGhvdXQgcHJvdmlkaW5nIGFueSBkZXRhaWxzLmBcbiAgICAgICAgICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgIHJlcGx5RXJyb3IobmV3IFJlc3BvbnNlRXJyb3I8dm9pZD4oRXJyb3JDb2Rlcy5NZXRob2ROb3RGb3VuZCwgYFVuaGFuZGxlZCBtZXRob2QgJHtyZXF1ZXN0TWVzc2FnZS5tZXRob2R9YCkpXG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBmdW5jdGlvbiBoYW5kbGVSZXNwb25zZShyZXNwb25zZU1lc3NhZ2U6IFJlc3BvbnNlTWVzc2FnZSk6IHZvaWQge1xuICAgICAgICBpZiAoaXNVbnN1YnNjcmliZWQoKSkge1xuICAgICAgICAgICAgLy8gU2VlIGhhbmRsZSByZXF1ZXN0LlxuICAgICAgICAgICAgcmV0dXJuXG4gICAgICAgIH1cblxuICAgICAgICBpZiAocmVzcG9uc2VNZXNzYWdlLmlkID09PSBudWxsKSB7XG4gICAgICAgICAgICBpZiAocmVzcG9uc2VNZXNzYWdlLmVycm9yKSB7XG4gICAgICAgICAgICAgICAgbG9nZ2VyLmVycm9yKFxuICAgICAgICAgICAgICAgICAgICBgUmVjZWl2ZWQgcmVzcG9uc2UgbWVzc2FnZSB3aXRob3V0IGlkOiBFcnJvciBpczogXFxuJHtKU09OLnN0cmluZ2lmeShcbiAgICAgICAgICAgICAgICAgICAgICAgIHJlc3BvbnNlTWVzc2FnZS5lcnJvcixcbiAgICAgICAgICAgICAgICAgICAgICAgIHVuZGVmaW5lZCxcbiAgICAgICAgICAgICAgICAgICAgICAgIDRcbiAgICAgICAgICAgICAgICAgICAgKX1gXG4gICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICBsb2dnZXIuZXJyb3IoYFJlY2VpdmVkIHJlc3BvbnNlIG1lc3NhZ2Ugd2l0aG91dCBpZC4gTm8gZnVydGhlciBlcnJvciBpbmZvcm1hdGlvbiBwcm92aWRlZC5gKVxuICAgICAgICAgICAgfVxuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgY29uc3Qga2V5ID0gU3RyaW5nKHJlc3BvbnNlTWVzc2FnZS5pZClcbiAgICAgICAgICAgIGNvbnN0IHJlc3BvbnNlUHJvbWlzZSA9IHJlc3BvbnNlUHJvbWlzZXNba2V5XVxuICAgICAgICAgICAgaWYgKHJlc3BvbnNlUHJvbWlzZSkge1xuICAgICAgICAgICAgICAgIHRyYWNlci5yZXNwb25zZVJlY2VpdmVkKFxuICAgICAgICAgICAgICAgICAgICByZXNwb25zZU1lc3NhZ2UsXG4gICAgICAgICAgICAgICAgICAgIHJlc3BvbnNlUHJvbWlzZS5yZXF1ZXN0IHx8IHJlc3BvbnNlUHJvbWlzZS5tZXRob2QsXG4gICAgICAgICAgICAgICAgICAgIHJlc3BvbnNlUHJvbWlzZS50aW1lclN0YXJ0XG4gICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgIGRlbGV0ZSByZXNwb25zZVByb21pc2VzW2tleV1cbiAgICAgICAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgICAgICAgICBpZiAocmVzcG9uc2VNZXNzYWdlLmVycm9yKSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBjb25zdCBlcnJvciA9IHJlc3BvbnNlTWVzc2FnZS5lcnJvclxuICAgICAgICAgICAgICAgICAgICAgICAgcmVzcG9uc2VQcm9taXNlLnJlamVjdChuZXcgUmVzcG9uc2VFcnJvcihlcnJvci5jb2RlLCBlcnJvci5tZXNzYWdlLCBlcnJvci5kYXRhKSlcbiAgICAgICAgICAgICAgICAgICAgfSBlbHNlIGlmIChyZXNwb25zZU1lc3NhZ2UucmVzdWx0ICE9PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICAgICAgICAgICAgICAgIHJlc3BvbnNlUHJvbWlzZS5yZXNvbHZlKHJlc3BvbnNlTWVzc2FnZS5yZXN1bHQpXG4gICAgICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgICAgICB0aHJvdyBuZXcgRXJyb3IoJ1Nob3VsZCBuZXZlciBoYXBwZW4uJylcbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGVycm9yKSB7XG4gICAgICAgICAgICAgICAgICAgIGlmIChlcnJvci5tZXNzYWdlKSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBsb2dnZXIuZXJyb3IoXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgYFJlc3BvbnNlIGhhbmRsZXIgJyR7cmVzcG9uc2VQcm9taXNlLm1ldGhvZH0nIGZhaWxlZCB3aXRoIG1lc3NhZ2U6ICR7ZXJyb3IubWVzc2FnZX1gXG4gICAgICAgICAgICAgICAgICAgICAgICApXG4gICAgICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBsb2dnZXIuZXJyb3IoYFJlc3BvbnNlIGhhbmRsZXIgJyR7cmVzcG9uc2VQcm9taXNlLm1ldGhvZH0nIGZhaWxlZCB1bmV4cGVjdGVkbHkuYClcbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgdHJhY2VyLnVua25vd25SZXNwb25zZVJlY2VpdmVkKHJlc3BvbnNlTWVzc2FnZSlcbiAgICAgICAgICAgIH1cbiAgICAgICAgfVxuICAgIH1cblxuICAgIGZ1bmN0aW9uIGhhbmRsZU5vdGlmaWNhdGlvbihtZXNzYWdlOiBOb3RpZmljYXRpb25NZXNzYWdlKTogdm9pZCB7XG4gICAgICAgIGlmIChpc1Vuc3Vic2NyaWJlZCgpKSB7XG4gICAgICAgICAgICAvLyBTZWUgaGFuZGxlIHJlcXVlc3QuXG4gICAgICAgICAgICByZXR1cm5cbiAgICAgICAgfVxuICAgICAgICBsZXQgbm90aWZpY2F0aW9uSGFuZGxlcjogR2VuZXJpY05vdGlmaWNhdGlvbkhhbmRsZXIgfCB1bmRlZmluZWRcbiAgICAgICAgaWYgKG1lc3NhZ2UubWV0aG9kID09PSBDYW5jZWxOb3RpZmljYXRpb24udHlwZSkge1xuICAgICAgICAgICAgbm90aWZpY2F0aW9uSGFuZGxlciA9IChwYXJhbXM6IENhbmNlbFBhcmFtcykgPT4ge1xuICAgICAgICAgICAgICAgIGNvbnN0IGlkID0gcGFyYW1zLmlkXG4gICAgICAgICAgICAgICAgY29uc3Qgc291cmNlID0gcmVxdWVzdFRva2Vuc1tTdHJpbmcoaWQpXVxuICAgICAgICAgICAgICAgIGlmIChzb3VyY2UpIHtcbiAgICAgICAgICAgICAgICAgICAgc291cmNlLmNhbmNlbCgpXG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgfVxuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgY29uc3QgZWxlbWVudCA9IG5vdGlmaWNhdGlvbkhhbmRsZXJzW21lc3NhZ2UubWV0aG9kXVxuICAgICAgICAgICAgaWYgKGVsZW1lbnQpIHtcbiAgICAgICAgICAgICAgICBub3RpZmljYXRpb25IYW5kbGVyID0gZWxlbWVudC5oYW5kbGVyXG4gICAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICAgICAgaWYgKG5vdGlmaWNhdGlvbkhhbmRsZXIgfHwgc3Rhck5vdGlmaWNhdGlvbkhhbmRsZXIpIHtcbiAgICAgICAgICAgIHRyeSB7XG4gICAgICAgICAgICAgICAgdHJhY2VyLm5vdGlmaWNhdGlvblJlY2VpdmVkKG1lc3NhZ2UpXG4gICAgICAgICAgICAgICAgbm90aWZpY2F0aW9uSGFuZGxlclxuICAgICAgICAgICAgICAgICAgICA/IG5vdGlmaWNhdGlvbkhhbmRsZXIobWVzc2FnZS5wYXJhbXMpXG4gICAgICAgICAgICAgICAgICAgIDogc3Rhck5vdGlmaWNhdGlvbkhhbmRsZXIhKG1lc3NhZ2UubWV0aG9kLCBtZXNzYWdlLnBhcmFtcylcbiAgICAgICAgICAgIH0gY2F0Y2ggKGVycm9yKSB7XG4gICAgICAgICAgICAgICAgaWYgKGVycm9yLm1lc3NhZ2UpIHtcbiAgICAgICAgICAgICAgICAgICAgbG9nZ2VyLmVycm9yKGBOb3RpZmljYXRpb24gaGFuZGxlciAnJHttZXNzYWdlLm1ldGhvZH0nIGZhaWxlZCB3aXRoIG1lc3NhZ2U6ICR7ZXJyb3IubWVzc2FnZX1gKVxuICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgIGxvZ2dlci5lcnJvcihgTm90aWZpY2F0aW9uIGhhbmRsZXIgJyR7bWVzc2FnZS5tZXRob2R9JyBmYWlsZWQgdW5leHBlY3RlZGx5LmApXG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgfVxuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgdW5oYW5kbGVkTm90aWZpY2F0aW9uRW1pdHRlci5maXJlKG1lc3NhZ2UpXG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBmdW5jdGlvbiBoYW5kbGVJbnZhbGlkTWVzc2FnZShtZXNzYWdlOiBNZXNzYWdlKTogdm9pZCB7XG4gICAgICAgIGlmICghbWVzc2FnZSkge1xuICAgICAgICAgICAgbG9nZ2VyLmVycm9yKCdSZWNlaXZlZCBlbXB0eSBtZXNzYWdlLicpXG4gICAgICAgICAgICByZXR1cm5cbiAgICAgICAgfVxuICAgICAgICBsb2dnZXIuZXJyb3IoXG4gICAgICAgICAgICBgUmVjZWl2ZWQgbWVzc2FnZSB3aGljaCBpcyBuZWl0aGVyIGEgcmVzcG9uc2Ugbm9yIGEgbm90aWZpY2F0aW9uIG1lc3NhZ2U6XFxuJHtKU09OLnN0cmluZ2lmeShcbiAgICAgICAgICAgICAgICBtZXNzYWdlLFxuICAgICAgICAgICAgICAgIG51bGwsXG4gICAgICAgICAgICAgICAgNFxuICAgICAgICAgICAgKX1gXG4gICAgICAgIClcbiAgICAgICAgLy8gVGVzdCB3aGV0aGVyIHdlIGZpbmQgYW4gaWQgdG8gcmVqZWN0IHRoZSBwcm9taXNlXG4gICAgICAgIGNvbnN0IHJlc3BvbnNlTWVzc2FnZTogUmVzcG9uc2VNZXNzYWdlID0gbWVzc2FnZSBhcyBSZXNwb25zZU1lc3NhZ2VcbiAgICAgICAgaWYgKHR5cGVvZiByZXNwb25zZU1lc3NhZ2UuaWQgPT09ICdzdHJpbmcnIHx8IHR5cGVvZiByZXNwb25zZU1lc3NhZ2UuaWQgPT09ICdudW1iZXInKSB7XG4gICAgICAgICAgICBjb25zdCBrZXkgPSBTdHJpbmcocmVzcG9uc2VNZXNzYWdlLmlkKVxuICAgICAgICAgICAgY29uc3QgcmVzcG9uc2VIYW5kbGVyID0gcmVzcG9uc2VQcm9taXNlc1trZXldXG4gICAgICAgICAgICBpZiAocmVzcG9uc2VIYW5kbGVyKSB7XG4gICAgICAgICAgICAgICAgcmVzcG9uc2VIYW5kbGVyLnJlamVjdChuZXcgRXJyb3IoJ1RoZSByZWNlaXZlZCByZXNwb25zZSBoYXMgbmVpdGhlciBhIHJlc3VsdCBub3IgYW4gZXJyb3IgcHJvcGVydHkuJykpXG4gICAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBmdW5jdGlvbiB0aHJvd0lmQ2xvc2VkT3JVbnN1YnNjcmliZWQoKTogdm9pZCB7XG4gICAgICAgIGlmIChpc0Nsb3NlZCgpKSB7XG4gICAgICAgICAgICB0aHJvdyBuZXcgQ29ubmVjdGlvbkVycm9yKENvbm5lY3Rpb25FcnJvcnMuQ2xvc2VkLCAnQ29ubmVjdGlvbiBpcyBjbG9zZWQuJylcbiAgICAgICAgfVxuICAgICAgICBpZiAoaXNVbnN1YnNjcmliZWQoKSkge1xuICAgICAgICAgICAgdGhyb3cgbmV3IENvbm5lY3Rpb25FcnJvcihDb25uZWN0aW9uRXJyb3JzLlVuc3Vic2NyaWJlZCwgJ0Nvbm5lY3Rpb24gaXMgdW5zdWJzY3JpYmVkLicpXG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBmdW5jdGlvbiB0aHJvd0lmTGlzdGVuaW5nKCk6IHZvaWQge1xuICAgICAgICBpZiAoaXNMaXN0ZW5pbmcoKSkge1xuICAgICAgICAgICAgdGhyb3cgbmV3IENvbm5lY3Rpb25FcnJvcihDb25uZWN0aW9uRXJyb3JzLkFscmVhZHlMaXN0ZW5pbmcsICdDb25uZWN0aW9uIGlzIGFscmVhZHkgbGlzdGVuaW5nJylcbiAgICAgICAgfVxuICAgIH1cblxuICAgIGZ1bmN0aW9uIHRocm93SWZOb3RMaXN0ZW5pbmcoKTogdm9pZCB7XG4gICAgICAgIGlmICghaXNMaXN0ZW5pbmcoKSkge1xuICAgICAgICAgICAgdGhyb3cgbmV3IEVycm9yKCdDYWxsIGxpc3RlbigpIGZpcnN0LicpXG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBjb25zdCBjb25uZWN0aW9uOiBDb25uZWN0aW9uID0ge1xuICAgICAgICBzZW5kTm90aWZpY2F0aW9uOiAobWV0aG9kOiBzdHJpbmcsIHBhcmFtczogYW55KTogdm9pZCA9PiB7XG4gICAgICAgICAgICB0aHJvd0lmQ2xvc2VkT3JVbnN1YnNjcmliZWQoKVxuICAgICAgICAgICAgY29uc3Qgbm90aWZpY2F0aW9uTWVzc2FnZTogTm90aWZpY2F0aW9uTWVzc2FnZSA9IHtcbiAgICAgICAgICAgICAgICBqc29ucnBjOiB2ZXJzaW9uLFxuICAgICAgICAgICAgICAgIG1ldGhvZCxcbiAgICAgICAgICAgICAgICBwYXJhbXMsXG4gICAgICAgICAgICB9XG4gICAgICAgICAgICB0cmFjZXIubm90aWZpY2F0aW9uU2VudChub3RpZmljYXRpb25NZXNzYWdlKVxuICAgICAgICAgICAgdHJhbnNwb3J0cy53cml0ZXIud3JpdGUobm90aWZpY2F0aW9uTWVzc2FnZSlcbiAgICAgICAgfSxcbiAgICAgICAgb25Ob3RpZmljYXRpb246ICh0eXBlOiBzdHJpbmcgfCBTdGFyTm90aWZpY2F0aW9uSGFuZGxlciwgaGFuZGxlcj86IEdlbmVyaWNOb3RpZmljYXRpb25IYW5kbGVyKTogdm9pZCA9PiB7XG4gICAgICAgICAgICB0aHJvd0lmQ2xvc2VkT3JVbnN1YnNjcmliZWQoKVxuICAgICAgICAgICAgaWYgKHR5cGVvZiB0eXBlID09PSAnZnVuY3Rpb24nKSB7XG4gICAgICAgICAgICAgICAgc3Rhck5vdGlmaWNhdGlvbkhhbmRsZXIgPSB0eXBlXG4gICAgICAgICAgICB9IGVsc2UgaWYgKGhhbmRsZXIpIHtcbiAgICAgICAgICAgICAgICBub3RpZmljYXRpb25IYW5kbGVyc1t0eXBlXSA9IHsgdHlwZTogdW5kZWZpbmVkLCBoYW5kbGVyIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgfSxcbiAgICAgICAgc2VuZFJlcXVlc3Q6IDxSPihtZXRob2Q6IHN0cmluZywgcGFyYW1zOiBhbnksIHRva2VuPzogQ2FuY2VsbGF0aW9uVG9rZW4pID0+IHtcbiAgICAgICAgICAgIHRocm93SWZDbG9zZWRPclVuc3Vic2NyaWJlZCgpXG4gICAgICAgICAgICB0aHJvd0lmTm90TGlzdGVuaW5nKClcbiAgICAgICAgICAgIHRva2VuID0gQ2FuY2VsbGF0aW9uVG9rZW4uaXModG9rZW4pID8gdG9rZW4gOiB1bmRlZmluZWRcbiAgICAgICAgICAgIGNvbnN0IGlkID0gc2VxdWVuY2VOdW1iZXIrK1xuICAgICAgICAgICAgY29uc3QgcmVzdWx0ID0gbmV3IFByb21pc2U8Uj4oKHJlc29sdmUsIHJlamVjdCkgPT4ge1xuICAgICAgICAgICAgICAgIGNvbnN0IHJlcXVlc3RNZXNzYWdlOiBSZXF1ZXN0TWVzc2FnZSA9IHtcbiAgICAgICAgICAgICAgICAgICAganNvbnJwYzogdmVyc2lvbixcbiAgICAgICAgICAgICAgICAgICAgaWQsXG4gICAgICAgICAgICAgICAgICAgIG1ldGhvZCxcbiAgICAgICAgICAgICAgICAgICAgcGFyYW1zLFxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgICBsZXQgcmVzcG9uc2VQcm9taXNlOiBSZXNwb25zZVByb21pc2UgfCBudWxsID0ge1xuICAgICAgICAgICAgICAgICAgICBtZXRob2QsXG4gICAgICAgICAgICAgICAgICAgIHJlcXVlc3Q6IHRyYWNlID09PSBUcmFjZS5WZXJib3NlID8gcmVxdWVzdE1lc3NhZ2UgOiB1bmRlZmluZWQsXG4gICAgICAgICAgICAgICAgICAgIHRpbWVyU3RhcnQ6IERhdGUubm93KCksXG4gICAgICAgICAgICAgICAgICAgIHJlc29sdmUsXG4gICAgICAgICAgICAgICAgICAgIHJlamVjdCxcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgdHJhY2VyLnJlcXVlc3RTZW50KHJlcXVlc3RNZXNzYWdlKVxuICAgICAgICAgICAgICAgIHRyeSB7XG4gICAgICAgICAgICAgICAgICAgIHRyYW5zcG9ydHMud3JpdGVyLndyaXRlKHJlcXVlc3RNZXNzYWdlKVxuICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgICAgICAgICAgICAgICAgLy8gV3JpdGluZyB0aGUgbWVzc2FnZSBmYWlsZWQuIFNvIHdlIG5lZWQgdG8gcmVqZWN0IHRoZSBwcm9taXNlLlxuICAgICAgICAgICAgICAgICAgICByZXNwb25zZVByb21pc2UucmVqZWN0KFxuICAgICAgICAgICAgICAgICAgICAgICAgbmV3IFJlc3BvbnNlRXJyb3I8dm9pZD4oRXJyb3JDb2Rlcy5NZXNzYWdlV3JpdGVFcnJvciwgZS5tZXNzYWdlID8gZS5tZXNzYWdlIDogJ1Vua25vd24gcmVhc29uJylcbiAgICAgICAgICAgICAgICAgICAgKVxuICAgICAgICAgICAgICAgICAgICByZXNwb25zZVByb21pc2UgPSBudWxsXG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIGlmIChyZXNwb25zZVByb21pc2UpIHtcbiAgICAgICAgICAgICAgICAgICAgcmVzcG9uc2VQcm9taXNlc1tTdHJpbmcoaWQpXSA9IHJlc3BvbnNlUHJvbWlzZVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH0pXG4gICAgICAgICAgICBpZiAodG9rZW4pIHtcbiAgICAgICAgICAgICAgICB0b2tlbi5vbkNhbmNlbGxhdGlvblJlcXVlc3RlZCgoKSA9PiB7XG4gICAgICAgICAgICAgICAgICAgIGNvbm5lY3Rpb24uc2VuZE5vdGlmaWNhdGlvbihDYW5jZWxOb3RpZmljYXRpb24udHlwZSwgeyBpZCB9KVxuICAgICAgICAgICAgICAgIH0pXG4gICAgICAgICAgICB9XG4gICAgICAgICAgICByZXR1cm4gcmVzdWx0XG4gICAgICAgIH0sXG4gICAgICAgIG9uUmVxdWVzdDogPFIsIEU+KHR5cGU6IHN0cmluZyB8IFN0YXJSZXF1ZXN0SGFuZGxlciwgaGFuZGxlcj86IEdlbmVyaWNSZXF1ZXN0SGFuZGxlcjxSLCBFPik6IHZvaWQgPT4ge1xuICAgICAgICAgICAgdGhyb3dJZkNsb3NlZE9yVW5zdWJzY3JpYmVkKClcblxuICAgICAgICAgICAgaWYgKHR5cGVvZiB0eXBlID09PSAnZnVuY3Rpb24nKSB7XG4gICAgICAgICAgICAgICAgc3RhclJlcXVlc3RIYW5kbGVyID0gdHlwZVxuICAgICAgICAgICAgfSBlbHNlIGlmIChoYW5kbGVyKSB7XG4gICAgICAgICAgICAgICAgcmVxdWVzdEhhbmRsZXJzW3R5cGVdID0geyB0eXBlOiB1bmRlZmluZWQsIGhhbmRsZXIgfVxuICAgICAgICAgICAgfVxuICAgICAgICB9LFxuICAgICAgICB0cmFjZTogKHZhbHVlOiBUcmFjZSwgX3RyYWNlcjogVHJhY2VyLCBzZW5kTm90aWZpY2F0aW9uID0gZmFsc2UpID0+IHtcbiAgICAgICAgICAgIHRyYWNlID0gdmFsdWVcbiAgICAgICAgICAgIGlmICh0cmFjZSA9PT0gVHJhY2UuT2ZmKSB7XG4gICAgICAgICAgICAgICAgdHJhY2VyID0gbm9vcFRyYWNlclxuICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICB0cmFjZXIgPSBfdHJhY2VyXG4gICAgICAgICAgICB9XG4gICAgICAgIH0sXG4gICAgICAgIG9uRXJyb3I6IGVycm9yRW1pdHRlci5ldmVudCxcbiAgICAgICAgb25DbG9zZTogY2xvc2VFbWl0dGVyLmV2ZW50LFxuICAgICAgICBvblVuaGFuZGxlZE5vdGlmaWNhdGlvbjogdW5oYW5kbGVkTm90aWZpY2F0aW9uRW1pdHRlci5ldmVudCxcbiAgICAgICAgb25VbnN1YnNjcmliZTogdW5zdWJzY3JpYmVFbWl0dGVyLmV2ZW50LFxuICAgICAgICB1bnN1YnNjcmliZTogKCkgPT4ge1xuICAgICAgICAgICAgaWYgKGlzVW5zdWJzY3JpYmVkKCkpIHtcbiAgICAgICAgICAgICAgICByZXR1cm5cbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIHN0YXRlID0gQ29ubmVjdGlvblN0YXRlLlVuc3Vic2NyaWJlZFxuICAgICAgICAgICAgdW5zdWJzY3JpYmVFbWl0dGVyLmZpcmUodW5kZWZpbmVkKVxuICAgICAgICAgICAgZm9yIChjb25zdCBrZXkgb2YgT2JqZWN0LmtleXMocmVzcG9uc2VQcm9taXNlcykpIHtcbiAgICAgICAgICAgICAgICByZXNwb25zZVByb21pc2VzW2tleV0ucmVqZWN0KFxuICAgICAgICAgICAgICAgICAgICBuZXcgQ29ubmVjdGlvbkVycm9yKFxuICAgICAgICAgICAgICAgICAgICAgICAgQ29ubmVjdGlvbkVycm9ycy5VbnN1YnNjcmliZWQsXG4gICAgICAgICAgICAgICAgICAgICAgICBgVGhlIHVuZGVybHlpbmcgSlNPTi1SUEMgY29ubmVjdGlvbiBnb3QgdW5zdWJzY3JpYmVkIHdoaWxlIHJlc3BvbmRpbmcgdG8gdGhpcyAke1xuICAgICAgICAgICAgICAgICAgICAgICAgICAgIHJlc3BvbnNlUHJvbWlzZXNba2V5XS5tZXRob2RcbiAgICAgICAgICAgICAgICAgICAgICAgIH0gcmVxdWVzdC5gXG4gICAgICAgICAgICAgICAgICAgIClcbiAgICAgICAgICAgICAgICApXG4gICAgICAgICAgICB9XG4gICAgICAgICAgICByZXNwb25zZVByb21pc2VzID0gT2JqZWN0LmNyZWF0ZShudWxsKVxuICAgICAgICAgICAgcmVxdWVzdFRva2VucyA9IE9iamVjdC5jcmVhdGUobnVsbClcbiAgICAgICAgICAgIG1lc3NhZ2VRdWV1ZSA9IG5ldyBMaW5rZWRNYXA8c3RyaW5nLCBNZXNzYWdlPigpXG4gICAgICAgICAgICB0cmFuc3BvcnRzLndyaXRlci51bnN1YnNjcmliZSgpXG4gICAgICAgICAgICB0cmFuc3BvcnRzLnJlYWRlci51bnN1YnNjcmliZSgpXG4gICAgICAgIH0sXG4gICAgICAgIGxpc3RlbjogKCkgPT4ge1xuICAgICAgICAgICAgdGhyb3dJZkNsb3NlZE9yVW5zdWJzY3JpYmVkKClcbiAgICAgICAgICAgIHRocm93SWZMaXN0ZW5pbmcoKVxuXG4gICAgICAgICAgICBzdGF0ZSA9IENvbm5lY3Rpb25TdGF0ZS5MaXN0ZW5pbmdcbiAgICAgICAgICAgIHRyYW5zcG9ydHMucmVhZGVyLmxpc3RlbihjYWxsYmFjaylcbiAgICAgICAgfSxcbiAgICB9XG5cbiAgICByZXR1cm4gY29ubmVjdGlvblxufVxuXG4vKiogU3VwcG9ydCBicm93c2VyIGFuZCBub2RlIGVudmlyb25tZW50cyB3aXRob3V0IG5lZWRpbmcgYSB0cmFuc3BpbGVyLiAqL1xuZnVuY3Rpb24gc2V0SW1tZWRpYXRlQ29tcGF0KGY6ICgpID0+IHZvaWQpOiB2b2lkIHtcbiAgICBpZiAodHlwZW9mIHNldEltbWVkaWF0ZSAhPT0gJ3VuZGVmaW5lZCcpIHtcbiAgICAgICAgc2V0SW1tZWRpYXRlKGYpXG4gICAgICAgIHJldHVyblxuICAgIH1cbiAgICBzZXRUaW1lb3V0KGYsIDApXG59XG4iXX0=