local grafana = import '/Users/efritz/Downloads/grafonnet-lib/grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;

dashboard.new(
  'LSIF Server WIP',
  schemaVersion=18,
  tags=[],
  editable=true,
)
.addPanel(
  graphPanel.new(
    title='Connection Cache Size',
  )
  .addTarget(
    prometheus.target(
      'lsif_connection_cache_size',
      legendFormat='Open Connections',
    )
  ),
  gridPos={
    x: 0,  y: 0,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Document Cache Size',
  )
  .addTarget(
    prometheus.target(
      'lsif_document_cache_size',
      legendFormat='Documents in Memory',
    )
  ),
  gridPos={
    x: 12, y: 0,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Connecton Cache Hit Ratio',
  )
  .addTargets([
    prometheus.target(
      'sum(rate(lsif_connection_cache_hit{type="hit"}[10m]))',
      legendFormat='Hit',
    ),
    prometheus.target(
      'sum(rate(lsif_connection_cache_hit{type="miss"}[10m]))',
      legendFormat='Miss',
    )
  ]),
  gridPos={
    x: 0,  y: 6,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Document Cache Hit Ratio',
  )
  .addTargets([
    prometheus.target(
      'sum(rate(lsif_document_cache_hit{type="hit"}[10m]))',
      legendFormat='Hit',
    ),
    prometheus.target(
      'sum(rate(lsif_document_cache_hit{type="miss"}[10m]))',
      legendFormat='Miss',
    )
  ]),
  gridPos={
    x: 12, y: 6,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Document Eviction Ratio',
  )
  .addTargets([
    prometheus.target(
      'sum(rate(lsif_connection_cache_eviction{type="evict"}[10m]))',
      legendFormat='Evict',
    ),
    prometheus.target(
      'sum(rate(lsif_connection_cache_eviction{type="locked"}[10m]))',
      legendFormat='Locked',
    )
  ]),
  gridPos={
    x: 0,  y: 12,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Document Eviction Ratio',
  )
  .addTargets([
    prometheus.target(
      'sum(rate(lsif_document_cache_eviction{type="evict"}[10m]))',
      legendFormat='Evict',
    ),
    prometheus.target(
      'sum(rate(lsif_document_cache_eviction{type="locked"}[10m]))',
      legendFormat='Locked',
    )
  ]),
  gridPos={
    x: 12, y: 12,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='HTTP Request Duration',
  )
  .addTarget(
    prometheus.target(
      'sum(rate(http_request_duration_seconds_bucket[10m])) by (le)',
      legendFormat='≤ {{le}}s',
      format='heatmap',
    )
  ),
  gridPos={
    x: 0,  y: 18,
    w: 24, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Database Query Duration',
  )
  .addTarget(
    prometheus.target(
      'sum(rate(lsif_database_query_duration_seconds_bucket[10m])) by (le)',
      legendFormat='≤ {{le}}s',
      format='heatmap',
    )
  ),
  gridPos={
    x: 0,  y: 24,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Database Insertion Duration',
  )
  .addTarget(
    prometheus.target(
      'sum(rate(lsif_database_insertion_duration_seconds_bucket[10m])) by (le)',
      legendFormat='≤ {{le}}s',
      format='heatmap',
    )
  ),
  gridPos={
    x: 12, y: 24,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Cross-Reop Database Query Duration',
  )
  .addTarget(
    prometheus.target(
      'sum(rate(lsif_xrepo_query_duration_seconds_bucket[10m])) by (le)',
      legendFormat='≤ {{le}}s',
      format='heatmap',
    )
  ),
  gridPos={
    x: 0,  y: 30,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Cross-Reop Database Insertion Duration',
  )
  .addTarget(
    prometheus.target(
      'sum(rate(lsif_xrepo_insertion_duration_seconds_bucket[10m])) by (le)',
      legendFormat='≤ {{le}}s',
      format='heatmap',
    )
  ),
  gridPos={
    x: 12, y: 30,
    w: 12, h: 6,
  }
)
.addPanel(
  graphPanel.new(
    title='Cross-Reop Database Query Duration',
  )
  .addTargets([
    prometheus.target(
      'sum(rate(lsif_bloom_filter_hit{type="hit"}[10m]))',
      legendFormat='Hit',
    ),
    prometheus.target(
      'sum(rate(lsif_bloom_filter_hit{type="miss"}[10m]))',
      legendFormat='Miss',
    ),
  ]),
  gridPos={
    x: 0,  y: 36,
    w: 12, h: 6,
  }
)
