from grafanalib.core import *

def service_qps_graph(name, job):
  return Graph(
    title='%s QPS' % (name,),
    dataSource="Prometheus",
    targets=[
      Target(
        expr=('sum(irate(src_%s_request_duration_seconds_count[1m]))' % (job,)),
        legendFormat="count",
        refId='A',
        ),
      Target(
        expr=('sum(irate(src_%s_request_duration_seconds_count{category="error"}[1m]))' % (job,)),
        legendFormat="error count",
        refId='B',
        ),      
      ],
    yAxes=[
      YAxis(format=OPS_FORMAT),
      YAxis(format=SHORT_FORMAT),
    ],
  )

def service_latency_graph(name, job):
  return Graph(
    title='%s Latency' % (name,),
    dataSource="Prometheus",
    targets=[
      Target(
        expr=('histogram_quantile(0.5, sum(irate(src_%s_request_duration_seconds_bucket[1m])) by (le))' % (job,)),
        legendFormat="0.5 quantile",
        refId='A',
      ),
      Target(
        expr=('histogram_quantile(0.99, sum(irate(src_%s_request_duration_seconds_bucket[1m])) by (le))' % (job,)),
        legendFormat="0.99 quantile",
        refId='B',
      ),
    ],
    yAxes=single_y_axis(format=SECONDS_FORMAT),
  )

def service_row(name, job):
  return Row(
    panels=[
      service_qps_graph(name, job),
      service_latency_graph(name, job),
    ]
  )
