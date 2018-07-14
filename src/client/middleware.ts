import { TextDocument } from 'vscode-languageserver-types'
import { NextSignature } from '../types/middleware'
import { HandleTextDocumentDecorationMiddleware, ProvideTextDocumentDecorationMiddleware } from './features/decoration'
import { ProvideTextDocumentHoverMiddleware } from './features/hover'

export interface Middleware {
    didOpen?: NextSignature<TextDocument, void>
    provideTextDocumentHover?: ProvideTextDocumentHoverMiddleware
    provideTextDocumentDecoration?: ProvideTextDocumentDecorationMiddleware
    handleTextDocumentDecoration?: HandleTextDocumentDecorationMiddleware
}
