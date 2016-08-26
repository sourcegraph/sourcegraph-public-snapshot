import Raven = require('..');

Raven.config('https://public@getsentry.com/1').install();

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
        'https://www.getsentry.com'
    ],
    whitelistUrls: [
        /https?:\/\/google\.com/,
        'https://www.google.com'
    ]
};

Raven.config('https://public@getsentry.com/1', options).install();

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

Raven.captureMessage('Broken!');
Raven.captureMessage('Broken!', {tags: { key: "value" }});

Raven.setDataCallback(function (data) {});
Raven.setDataCallback(function (data, original) {});
Raven.setShouldSendCallback(function (data) {});
Raven.setShouldSendCallback(function (data, original) {});
