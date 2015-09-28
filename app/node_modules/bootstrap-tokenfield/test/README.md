Testing
=======

The tests have been set up to run with a simulated DOM created with JSDOM.
Mouse and keyboard events are simulated using jQuery and jquery-simulate-ext plugin.

We looked into several other alternatives like Zombie.js and Karma, but
for our use case, they did not seem to offer many advantages over a JSDOM.
The main reason behind this is that neither of those options provide a
built-in or better way to simulate mouse and keyboard events.

A step up from the current test solution would be to create Selenium or PtahnomJS
tests so that we do not have to simulate mouse and keyboard, but could actually
control the browser. This will be on the roadmap in the future.

For now - simply do `npm test`.

### Writing tests

Take care to follow the style of existing tests. All tests that require DOM
manipulation and/or involve simulating user interaction, should go under
integration tests. Testing tokenfield methods should go under unit tests.
