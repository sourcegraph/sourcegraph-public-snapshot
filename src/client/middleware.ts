import { TextDocumentItem } from 'vscode-languageserver-types'
import { ReferenceParams } from '../protocol'
import { NextSignature } from '../types/middleware'
import { ProvideTextDocumentDecorationMiddleware } from './features/decoration'
import { ProvideTextDocumentHoverMiddleware } from './features/hover'
import { ProvideTextDocumentLocationMiddleware } from './features/location'

export interface Middleware {
    didOpen?: NextSignature<TextDocumentItem, void>
    didClose?: NextSignature<TextDocumentItem, void>
    provideTextDocumentDefinition?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentImplementation?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentReferences?: ProvideTextDocumentLocationMiddleware<ReferenceParams>
    provideTextDocumentTypeDefinition?: ProvideTextDocumentLocationMiddleware
    provideTextDocumentHover?: ProvideTextDocumentHoverMiddleware
    provideTextDocumentDecoration?: ProvideTextDocumentDecorationMiddleware
}
