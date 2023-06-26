import { CurrentDocumentContextWithLanguage, Completion } from '@sourcegraph/cody-shared/src/autocomplete'
import { ManualCompletionService } from '@sourcegraph/cody-shared/src/autocomplete/manual'

export class ManualCompletionServiceAgent extends ManualCompletionService {
    getCurrentDocumentContext(): Promise<CurrentDocumentContextWithLanguage | null> {
        throw new Error('Method not implemented.')
    }

    emitCompletions(docContext: CurrentDocumentContextWithLanguage, completions: Completion[]): void {
        throw new Error('Method not implemented.')
    }
}
