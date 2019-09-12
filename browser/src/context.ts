enum AppEnv {
    Extension,
    Page,
}

enum ScriptEnv {
    Content,
    Background,
    Options,
}

interface AppContext {
    appEnv: AppEnv
    scriptEnv: ScriptEnv
}

function getContext(): AppContext {
    const appEnv = window.SG_ENV === 'EXTENSION' ? AppEnv.Extension : AppEnv.Page

    let scriptEnv: ScriptEnv = ScriptEnv.Content
    if (appEnv === AppEnv.Extension) {
        if (window.location.pathname.includes('options.html')) {
            scriptEnv = ScriptEnv.Options
        } else if (globalThis.browser && browser.runtime.getBackgroundPage) {
            scriptEnv = ScriptEnv.Background
        }
    }

    return {
        appEnv,
        scriptEnv,
    }
}

const ctx = getContext()

export const isBackground = ctx.scriptEnv === ScriptEnv.Background
export const isOptions = ctx.scriptEnv === ScriptEnv.Options

export const isExtension = ctx.appEnv === AppEnv.Extension
export const isInPage = !isExtension

export const isPhabricator = Boolean(document.querySelector('.phabricator-wordmark'))
