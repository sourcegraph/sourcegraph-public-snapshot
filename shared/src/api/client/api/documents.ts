import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { from, Subscription, Unsubscribable } from 'rxjs'
import { concatMap } from 'rxjs/operators'
import { getModeFromPath } from '../../../languages'
import { ExtDocumentsAPI } from '../../extension/api/documents'
import { FileSystemService } from '../services/fileSystemService'
import { ModelService, TextModel } from '../services/modelService'

/** @internal */
export interface ClientDocumentsAPI extends ProxyValue {
    $openTextDocument(uri: string): Promise<TextModel>
}

/** @internal */
export class ClientDocuments implements ClientDocumentsAPI, Unsubscribable {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()

    constructor(
        proxy: ProxyResult<ExtDocumentsAPI>,
        private fileSystemService: Pick<FileSystemService, 'readFile'>,
        private modelService: ModelService
    ) {
        this.subscriptions.add(
            from(modelService.models)
                .pipe(concatMap(models => proxy.$acceptDocumentData(models)))
                .subscribe()
        )
    }

    public async $openTextDocument(uri: string): Promise<TextModel> {
        const model: TextModel = {
            uri,
            languageId: getModeFromPath(uri),
            text: await this.fileSystemService.readFile(new URL(uri)),
        }
        if (!this.modelService.hasModel(uri)) {
            this.modelService.addModel(model)
        }
        return model
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
