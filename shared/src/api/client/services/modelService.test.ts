import { Observable } from 'rxjs'
import { tap } from 'rxjs/operators'
import { createModelService, ModelService, TextModelUpdate, TextModel } from './modelService'

export function createTestModelService({
    models,
    updates,
}: {
    models?: TextModel[]
    updates?: Observable<TextModelUpdate[]>
}): ModelService {
    const service = createModelService()
    if (models) {
        for (const m of models) {
            service.addModel(m)
        }
    }
    const modelUpdates = updates
        ? updates.pipe(
              tap(updates => {
                  for (const update of updates) {
                      switch (update.type) {
                          case 'added':
                              service.addModel(update)
                              break
                          case 'updated':
                              service.updateModel(update.uri, update.text)
                              break
                          case 'deleted':
                              service.removeModel(update.uri)
                              break
                      }
                  }
              })
          )
        : service.modelUpdates
    return {
        ...service,
        modelUpdates,
    }
}

describe('ModelService', () => {
    describe('addModel', () => {
        it('adds', () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            expect([...modelService.models.values()]).toEqual([
                {
                    uri: 'u',
                    text: 't',
                    languageId: 'l',
                },
            ])
        })
        it('refuses to add model with duplicate URI', () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            expect(() => {
                modelService.addModel({ uri: 'u', text: 't2', languageId: 'l2' })
            }).toThrowError('model already exists with URI u')
            expect([...modelService.models.values()]).toEqual([
                {
                    uri: 'u',
                    text: 't',
                    languageId: 'l',
                },
            ])
        })
    })

    test('hasModel', () => {
        const modelService = createModelService()
        modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
        expect(modelService.hasModel('u')).toBeTruthy()
        expect(modelService.hasModel('u2')).toBeFalsy()
    })

    describe('updateModel', () => {
        test('existing model', () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            modelService.updateModel('u', 't2')
            expect([...modelService.models.values()]).toEqual([{ uri: 'u', text: 't2', languageId: 'l' }])
        })

        test('nonexistent model', () => {
            const modelService = createModelService()
            expect(() => modelService.updateModel('x', 't2')).toThrowError('model does not exist with URI x')
        })
    })

    describe('removeModel', () => {
        test('removes', () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            modelService.addModel({ uri: 'u2', text: 't2', languageId: 'l2' })
            modelService.removeModel('u')
            expect([...modelService.models.values()]).toEqual([
                {
                    uri: 'u2',
                    text: 't2',
                    languageId: 'l2',
                },
            ])
        })
    })
})
