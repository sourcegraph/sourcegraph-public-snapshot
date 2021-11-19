# How to guides

- [How to debug live code](debug_live_code.md)
- [Set up local development with Zoekt and Sourcegraph](zoekt_local_dev.md)
- [Ignoring editor config files in Git](ignoring_editor_config_files.md)
- [Use `golangci-lint`](use-golangci-lint.md)
- [Set up local Sourcegraph monitoring development](monitoring_local_dev.md)

## Offline development

Sometimes you will want to develop Sourcegraph but it just so happens you will be on a plane or a
train or perhaps a beach, and you will have no WiFi. And you may raise your fist toward heaven and
say something like, "Why, we can put a man on the moon, so why can't we develop high-quality code
search without an Internet connection?" But lower your hand back to your keyboard and fret no
further, you *can* develop Sourcegraph with no connectivity by setting the
`OFFLINE` environment variable:

```bash
OFFLINE=true sg start
```
