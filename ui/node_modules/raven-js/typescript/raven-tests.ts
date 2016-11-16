import Raven = require('..');

Raven.config('https://public@sentry.io/1').install();

var options = {
    logger: 'my-logger',
    ignoreUrls: [
        /graph\.facebook\.com/i,
        'graph.facebook.com'
    ],
    ignoreErrors: [
        /fb_xd_fragment/,
        'fb_xd_fragment'
    ],
    includePaths: [
        /https?:\/\/(www\.)?getsentry\.com/,
        'https://www.sentry.io'
    ],
    whitelistUrls: [
        /https?:\/\/google\.com/,
        'https://www.google.com'
    ]
};

Raven.config('https://public@sentry.io/1', options).install();

var throwsError = () => {
    throw new Error('broken');
};

try {
    throwsError();
} catch(e) {
    Raven.captureException(e);
    Raven.captureException(e, {tags: { key: "value" }});
}

Raven.context(throwsError);
Raven.context({tags: { key: "value" }}, throwsError);

setTimeout(Raven.wrap(throwsError), 1000);
Raven.wrap({logger: "my.module"}, throwsError)();

Raven.setUserContext({
    email: 'matt@example.com',
    id: '123'
});

Raven.setExtraContext({foo: 'bar'});
Raven.setTagsContext({env: 'prod'});
Raven.clearContext();
var obj:Object = Raven.getContext();
var err:Error = Raven.lastException();

Raven.captureMessage('Broken!');
Raven.captureMessage('Broken!', {tags: { key: "value" }});
+Raven.captureMessage('Broken!', { stacktrace: true });
Raven.captureBreadcrumb({});

Raven.setRelease('abc123');
Raven.setEnvironment('production');

Raven.setDataCallback(function (data: any) {});
Raven.setDataCallback(function (data: any, original: any) {});
Raven.setShouldSendCallback(function (data: any) {});
Raven.setShouldSendCallback(function (data: any, original: any) {});

Raven.showReportDialog({
    eventId: 'abcdef123456'
});
