# How to guides

- [Configure a test instance of Phabricator and Gitolite](configure_phabricator_gitolite.md)
- [Test a Phabricator and Gitolite instance](test_phabricator.md)

## Helpful how-tos

- [How to run tests](../concepts/testing.md)
- [How to debug live code](debug_live_code.md)
- [Windows support](#windows-support)
- [Other nice things](#other-nice-things)
  - [Offline development](#offline-development)

## Windows support

Running Sourcegraph on Windows is not actively tested, but should be possible within the Windows Subsystem for Linux (WSL).
Sourcegraph currently relies on Unix specifics in several places, which makes it currently not possible to run Sourcegraph directly inside Windows without WSL.
We are happy to accept contributions here! :)

## Other nice things

### Offline development

Sometimes you will want to develop Sourcegraph but it just so happens you will be on a plane or a
train or perhaps a beach, and you will have no WiFi. And you may raise your fist toward heaven and
say something like, "Why, we can put a man on the moon, so why can't we develop high-quality code
search without an Internet connection?" But lower your hand back to your keyboard and fret no
further, for the year is 2019, and you *can* develop Sourcegraph with no connectivity by setting the
`OFFLINE` environment variable:

```bash
OFFLINE=true dev/start.sh
```
