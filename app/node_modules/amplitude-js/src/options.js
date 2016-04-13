var language = require('./language');

// default options
module.exports = {
  apiEndpoint: 'api.amplitude.com',
  cookieExpiration: 365 * 10,
  cookieName: 'amplitude_id',
  domain: undefined,
  includeUtm: false,
  language: language.language,
  optOut: false,
  platform: 'Web',
  savedMaxCount: 1000,
  saveEvents: true,
  sessionTimeout: 30 * 60 * 1000,
  unsentKey: 'amplitude_unsent',
  unsentIdentifyKey: 'amplitude_unsent_identify',
  uploadBatchSize: 100,
  batchEvents: false,
  eventUploadThreshold: 30,
  eventUploadPeriodMillis: 30 * 1000, // 30s
};
