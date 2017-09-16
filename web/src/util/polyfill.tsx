import { URL, URLSearchParams } from 'whatwg-url'

const GLOBAL = global as any
GLOBAL.URL = URL
GLOBAL.URLSearchParams = URLSearchParams
