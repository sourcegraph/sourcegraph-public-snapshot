# Automation Store

Generate `automation_schema.sql` with this:

```bash
pg_dump sourcegraph -t "changeset*" -t "campaign*" -s |\
  ruby -e 'puts ARGF.read.split("\n\n").select {|l| l.start_with?("CREATE TABLE") && !(l.include?("citext") || l.include?("critical_or_site"))}'\
  > automation_schema.sql
```
