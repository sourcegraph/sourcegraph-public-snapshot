import { of } from 'rxjs'
import { EMPTY_ENVIRONMENT } from './environment'
import { Services } from './services'

describe('Services', () => {
    it('initializes empty services', () => {
        // tslint:disable-next-line:no-unused-expression
        new Services(of(EMPTY_ENVIRONMENT))
    })
})
