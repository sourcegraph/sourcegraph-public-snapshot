import { from } from 'rxjs'
import { first } from 'rxjs/operators'
import { createModelService } from './modelService'

describe('ModelService', () => {
    describe('addModel', () => {
        it('adds', async () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            expect(
                await from(modelService.models)
                    .pipe(first())
                    .toPromise()
            ).toEqual([
                {
                    uri: 'u',
                    text: 't',
                    languageId: 'l',
                },
            ])
        })
        it('refuses to add model with duplicate URI', async () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            expect(() => {
                modelService.addModel({ uri: 'u', text: 't2', languageId: 'l2' })
            }).toThrowError('model already exists with URI u')
            expect(
                await from(modelService.models)
                    .pipe(first())
                    .toPromise()
            ).toEqual([
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

    describe('removeModel', () => {
        test('removes', async () => {
            const modelService = createModelService()
            modelService.addModel({ uri: 'u', text: 't', languageId: 'l' })
            modelService.addModel({ uri: 'u2', text: 't2', languageId: 'l2' })
            modelService.removeModel('u')
            expect(
                await from(modelService.models)
                    .pipe(first())
                    .toPromise()
            ).toEqual([
                {
                    uri: 'u2',
                    text: 't2',
                    languageId: 'l2',
                },
            ])
        })
        test('noop if model not found', () => {
            const modelService = createModelService()
            modelService.removeModel('x')
        })
    })
})
