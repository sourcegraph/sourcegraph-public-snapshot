export default function checkReporter(reporter) {
  try {
    require(`mocha/lib/reporters/${reporter}`);
  } catch (errModule) {
    try {
      require(reporter);
    } catch (errLocal) {
      throw new Error(`reporter "${reporter}" does not exist`);
    }
  }
}
