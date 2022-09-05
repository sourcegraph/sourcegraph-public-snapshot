import { ROOT_PATH } from '@sourcegraph/build-config'
import { BUILD_OPTIONS as WEB_BUILD_OPTIONS } from '@sourcegraph/web/dev/esbuild/build'
import * as esbuild from 'esbuild'
import path from 'path'

const consoleWebRootPath = path.join(ROOT_PATH, 'enterprise/cmd/console/web')

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    ...WEB_BUILD_OPTIONS,
    entryPoints: { index: path.join(consoleWebRootPath, 'src/index.tsx') },
    outdir: path.join(consoleWebRootPath, 'static/dist'),
}
