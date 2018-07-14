import { TextDocument } from 'vscode-languageserver-types'
import { NextSignature } from '../types/middleware'
import { HandleTextDocumentDecorationMiddleware, ProvideTextDocumentDecorationMiddleware } from './features/decoration'
import { ProvideTextDocumentDefinitionMiddleware } from './features/definition'
import { ProvideTextDocumentHoverMiddleware } from './features/hover'

export interface Middleware {
    didOpen?: NextSignature<TextDocument, void>
    provideTextDocumentDefinition?: ProvideTextDocumentDefinitionMiddleware
    provideTextDocumentHover?: ProvideTextDocumentHoverMiddleware
    provideTextDocumentDecoration?: ProvideTextDocumentDecorationMiddleware
    handleTextDocumentDecoration?: HandleTextDocumentDecorationMiddleware
}
