## Sourcegraph Frontend Platform Prototype Refinement Bot

### Description:

Prototype Github Action to allow Frontend Platorm to refine async style. Works by utilizing github actions to send messages in slack combined with a firebase function to receive responses from the slack bot and update github accordingly.

This concept is still very much in Alpha 0 and is being tested with Frontend platform only. Feel free to reach out to #frontend-platform if there are any issues

### Structure

- github actions refinement-notify.yml
  - slack refinement notify
  - triggers on workflow dispatch and new issues being labeled or reopend
  - Outbounds scripts (sending to slack)
    - SendMessage.ts => sends actual slack message
    - GetRandomIssue.ts => Finds a random issue to pull from to be refined
  - Inbound Scripts (receiving from slack)
    - Utilizes refinement-bot/actions
    - Leverages a firebase function to issue responses
