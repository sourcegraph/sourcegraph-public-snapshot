import strict_rfc3339
import time
from spanid import *

__schemaPrefix = "_schema:"

# timeString returns a appdash-compatable (RFC 3339 / UTC offset) time string
# for the given timestamp (float seconds since epoch).
def timeString(ts):
    return strict_rfc3339.timestamp_to_rfc3339_utcoffset(ts)

# MarshalEvent marshals an event into annotations.
def MarshalEvent(e):
    a = []
    for key, value in e.marshal().items():
        a.append(Annotation(key, value))
    a.append(Annotation(__schemaPrefix + e.schema, ""))
    return a

# SpanNameEvent is an event which sets a span's name.
class SpanNameEvent:
    schema = "name"

    # name is literally the span's name.
    name = ""

    def __init__(self, name):
        self.name = name

    # marshal returns a dictionary of this event's value by name.
    def marshal(self):
        return {"Name": self.name}

# MsgEvent is an event that contains only a human-readable message.
class MsgEvent:
    schema = "msg"

    # msg is literally the message string.
    msg = ""

    def __init__(self, msg):
        self.msg = msg

    # marshal returns a dictionary of this event's value by schema.
    def marshal(self):
        return {"Msg": self.msg}

# LogEvent is an event whose timestamp is the current time and contains the
# given human-readable log message.
class LogEvent:
    schema = "log"

    # msg is literally the message string.
    msg = ""

    # RFC3339-UTC timestamp of the event.
    time = ""

    def __init__(self, msg):
        self.msg = msg
        self.time = timeString(time.time())

    def marshal(self):
        return {"Msg": self.msg, "Time": self.time}

# SQLEvent is an SQL query event with send and receive times, as well as the
# actual SQL that was ran, and a optional tag.
class SQLEvent:
    schema = "SQL"

    # sql is literally the SQL query that was ran.
    sql = ""

    # tag is a optional user-created tag associated with the SQL event. 
    tag = ""

    # RFC3339-UTC timestamp of when the query was sent, and later a result received.
    clientSend = ""
    clientRecv = ""

    def __init__(self, sql, send, recv=None, tag=""):
        self.sql = sql
        self.tag = tag
        self.clientSend = timeString(send)

        # If user didn't specify a recv time, use right now.
        if recv:
            self.clientRecv = timeString(recv)
        else:
            self.clientRecv = timeString(time.time())

    def marshal(self):
        return {
            "SQL": self.sql,
            "Tag": self.tag,
            "ClientSend": self.clientSend,
            "ClientRecv": self.clientRecv,
        }

# _flattenReqRes flattens the request or response info dictionary. It has very
# defined functionality primarily to avoid code duplication four times below in
# marshaling of HTTP events.
def _flattenReqRes(reqRes, into):
    for k, v in reqRes.items():
        if k == "Headers":
            into["Request.Headers." + k] = str(v)
        else:
            into["Request." + k] = str(v)
    return into

# HTTPServerEvent represents a HTTP event where a client's request was served.
#
#  e = HTTPServerEvent(
#      send = sendTimestamp,
#      recv = recvTimestamp,
#      route = "/endpoint-B",
#      user = "",
#      requestInfo = {
#          "Method": "GET",
#          "URI": "/endpoint-B",
#          "Proto": "HTTP/1.1",
#          "Headers": {}, # dictionary of strings in proper HTTP case "Foo-Bar".
#          "Host": "localhost:8699",
#          "RemoteAddr": "127.0.0.1:35787",
#          "ContentLength": 0,
#      },
#      responseInfo = {
#          "Method": "GET",
#          "URI": "/endpoint-B",
#          "Proto": "HTTP/1.1",
#          "Headers": { # dictionary of strings in proper HTTP case "Foo-Bar".
#              "Span-Id": "8d4bdb285382e850/ef1ba6f3fa12d5d2/d9564f71aae8aba2",
#          },
#          "Host": "localhost:8699",
#          "RemoteAddr": "127.0.0.1:35787",
#          "ContentLength": 28,
#      },
#  )
#
class HTTPServerEvent:
    schema = "HTTPServer"

    # Request headers and information.
    requestInfo = {}

    # Response headers and information.
    responseInfo = {}

    # The route, e.g. "/endpoint-B".
    route = ""

    # The user (if any).
    user = ""

    # RFC3339-UTC timestamp of when the query was sent, and later a result received.
    serverRecv = ""
    serverSend = ""

    def __init__(self, send, recv=None, requestInfo = {}, responseInfo = {}, route = "", user=""):
        self.requestInfo = requestInfo
        self.responseInfo = responseInfo
        self.route = route
        self.user = user
        self.serverSend = timeString(send)

        # If user didn't specify a recv time, use right now.
        if recv:
            self.serverRecv = timeString(recv)
        else:
            self.serverRecv = timeString(time.time())

    def marshal(self):
        d = {
            "Route": self.route,
            "User": self.user,
            "ServerRecv": self.serverRecv,
            "ServerSend": self.serverSend,
        }
        d = _flattenReqRes(self.requestInfo, d)
        d = _flattenReqRes(self.responseInfo, d)
        return d

# HTTPClientEvent represents a HTTP event where a outbound request to a server
# was made.
#
#  e = HTTPClientEvent(
#      send = sendTimestamp,
#      recv = recvTimestamp,
#      requestInfo = {
#          "Method": "GET",
#          "URI": "/endpoint-B",
#          "Proto": "HTTP/1.1",
#          "Headers": {}, # dictionary of strings in proper HTTP case "Foo-Bar".
#          "Host": "localhost:8699",
#          "RemoteAddr": "127.0.0.1:35787",
#          "ContentLength": 0,
#      },
#      responseInfo = {
#          "Method": "GET",
#          "URI": "/endpoint-B",
#          "Proto": "HTTP/1.1",
#          "Headers": { # dictionary of strings in proper HTTP case "Foo-Bar".
#              "Span-Id": "8d4bdb285382e850/ef1ba6f3fa12d5d2/d9564f71aae8aba2",
#          },
#          "Host": "localhost:8699",
#          "RemoteAddr": "127.0.0.1:35787",
#          "ContentLength": 28,
#      },
#  )
#
class HTTPClientEvent:
    schema = "HTTPClient"

    # Request headers and information.
    requestInfo = {}

    # Response headers and information.
    responseInfo = {}

    # RFC3339-UTC timestamp of when the query was sent, and later a result received.
    serverRecv = ""
    serverSend = ""

    def __init__(self, send, recv=None, requestInfo = {}, responseInfo = {}):
        self.requestInfo = requestInfo
        self.responseInfo = responseInfo
        self.serverSend = timeString(send)

        # If user didn't specify a recv time, use right now.
        if recv:
            self.serverRecv = timeString(recv)
        else:
            self.serverRecv = timeString(time.time())

    def marshal(self):
        d = {
            "ServerRecv": self.serverRecv,
            "ServerSend": self.serverSend,
        }
        d = _flattenReqRes(self.requestInfo, d)
        d = _flattenReqRes(self.responseInfo, d)
        return d

