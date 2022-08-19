import { cleanEnv, str } from 'envalid'
import Libhoney from 'libhoney'

const environment = cleanEnv(process.env, {
    HONEYCOMB_API_KEY: str(),
    HONEYCOMB_DATASET: str({ default: 'client-infrastructure', choices: ['client-infrastructure'] }),
})

export const libhoneySDK = new Libhoney({
    writeKey: environment.HONEYCOMB_API_KEY,
    dataset: environment.HONEYCOMB_DATASET,
})
