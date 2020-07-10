# Zoekt

TBD

## Config

TBD

## Diagram

```mermaid
graph TD
  http --> frontend
  graphql --> frontend
  serviceN[Other Services] --> frontend
  frontend --> db[(PostgreSQL)]
  frontend --> cache[(PostgreSQL)]
```
