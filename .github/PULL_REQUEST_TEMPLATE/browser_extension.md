


#### Test Plan

##### Code Hosts

- [ ] GitHub
- [ ] GitHub Enterprise
- [ ] Refined GitHub
- [ ] Phabricator
    - [ ] Navigate to https://phabricator.sgdev.org/config/edit/sourcegraph.enabled/ and set sourcegraph.enabled to `false`
    - [ ] Point your browser extension to a Sourcegraph instance with the following external service:
        ```
        {
          "prefix": "gitolite.sgdev.org/",
          "host": "git@gitolite.sgdev.org",
        }
        ```
    - [ ] Verify the callsign mappings are correctly set at https://phabricator.sgdev.org/config/edit/sourcegraph.callsignMappings/
    - [ ] Make sure your browser extension has permissions for `https://phabricator.sgdev.org` (you can check this through the popup)
    - [ ] Navigate to a single file: https://phabricator.sgdev.org/source/test/browse/master/main.go
        - [ ] Verify "View on Sourcegraph" button is present and working correctly
        - [ ] Verify hovers work as expected
    - [ ] Navigate to a diff: https://phabricator.sgdev.org/D3
        - [ ] Verify "View on Sourcegraph" buttons are present on all change types, and working correctly
        - [ ] Verify hovers are working correctly on added, removed, unchanged lines
- [ ] Phabricator integration
    - [ ] Run a local Sourcegraph dev instance tunnelled through ngrok
    - [ ] Set `corsOrigin` to `"https://phabricator.sgdev.org"` in your site config
    - [ ] Add the following Gitolite external service:
        ```
        {
          "prefix": "gitolite.sgdev.org/",
          "host": "git@gitolite.sgdev.org",
        }
        ```
    - [ ] Verify that the phabricator assets are served:
        - [ ] %NGROK_URL%/.assets/extension/scripts/phabricator.bundle.js
        - [ ] %NGROK_URL%/.assets/extension/css/style.bundle.css
    - [ ] Navigate to https://phabricator.sgdev.org/config/edit/sourcegraph.url/ and set the Sourcegraph URL to your tunnelled ngrok URL
    - [ ] Verify the callsign mappings are correctly set at https://phabricator.sgdev.org/config/edit/sourcegraph.callsignMappings/
    - [ ] Navigate to a single file: https://phabricator.sgdev.org/source/test/browse/master/main.go
        - [ ] Verify "View on Sourcegraph" button is present and working correctly
        - [ ] Verify hovers work as expected
    - [ ] Navigate to a diff: https://phabricator.sgdev.org/D3
        - [ ] Verify "View on Sourcegraph" buttons are present on all change types, and working correctly
        - [ ] Verify hovers are working correctly on added, removed, unchanged lines
- [ ] Bitbucket
- [ ] Gitlab

##### Browsers

- [ ] Chrome
- [ ] Firefox
