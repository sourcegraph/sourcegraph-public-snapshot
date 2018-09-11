import { EMPTY_OBSERVABLE_ENVIRONMENT } from './environment'
import { Registries } from './registries'

describe('Registries', () => {
    it('initializes empty registries', () => {
        // tslint:disable-next-line:no-unused-expression
        new Registries(EMPTY_OBSERVABLE_ENVIRONMENT)
    })
})
