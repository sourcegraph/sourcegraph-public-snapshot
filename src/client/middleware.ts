import { TextDocument } from 'vscode-languageserver-types'
import { NextSignature } from '../types/middleware'
import { HandleTextDocumentDecorationMiddleware, ProvideTextDocumentDecorationMiddleware } from './features/decoration'
import { ProvideTextDocumentHoverMiddleware } from './features/hover'
import { ProvideTextDocumentLocationMiddleware } from './features/location'

export interface Middleware {
    didOpen?: NextSignature<TextDocument, void>
    provideTextDocumentDefinition?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentImplementation?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentTypeDefinition?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentHover?: ProvideTextDocumentHoverMiddleware
    provideTextDocumentDecoration?: ProvideTextDocumentDecorationMiddleware
    handleTextDocumentDecoration?: HandleTextDocumentDecorationMiddleware
}
