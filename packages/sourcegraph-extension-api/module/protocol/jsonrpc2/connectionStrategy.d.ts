import { Message, ResponseMessage } from './messages';
export interface ConnectionStrategy {
    cancelUndispatched?: (message: Message, next: (message: Message) => ResponseMessage | undefined) => ResponseMessage | undefined;
}
