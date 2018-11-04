import { of } from 'rxjs'
import { EMPTY_ENVIRONMENT } from './environment'
import { Registries } from './registries'

describe('Registries', () => {
    it('initializes empty registries', () => {
        // tslint:disable-next-line:no-unused-expression
        new Registries(of(EMPTY_ENVIRONMENT))
    })
})
