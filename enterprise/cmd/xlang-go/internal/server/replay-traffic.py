#!/usr/bin/env python

# This script takes the Debug log output from the golang build server, and
# will play it back to servers via sourcegraph.com's LSP HTTP layer

import json
import requests

api = 'http://127.0.0.1/.api/xlang/'
# Uncomment for production
#api = 'https://sourcegraph.com/.api/xlang/'

def main(lines):
    for num, line in enumerate(lines):
	line = line.strip()
	i = line.find('>>> ')
	if i < 0:
	    continue
	root_uri, jsonrpc_id, method, params = line[i + 4:].split(' ', 3)
	if method in ('initialize', 'shutdown', 'exit'):
	    continue
	print(line[i + 4:])
	data = [
	    {"id":0,"method":"initialize","params":{"rootUri":root_uri,"mode":"go"}},
	    {"id":1,"method":method,"params":json.loads(params.replace("file:///", root_uri + "#"))},
	    {"id":2,"method":"shutdown"},
	    {"method":"exit"},
	]
	r = requests.post(api + method, json=data)
	print(r.status_code, r.reason)

if __name__ == '__main__':
    import fileinput
    main(fileinput.input())

run_tests = False
if run_tests:
    main('''2017-01-19T20:53:54.986751588Z >>> git://github.com/captncraig/mux?acfc892941192f90aadd4f452a295bf39fc5f7ed 229 textDocument/hover {"position":{"character":5,"line":23},"textDocument":{"uri":"file:///mux.go"}}
2017-01-19T20:53:54.989759007Z <<< git://github.com/captncraig/mux?acfc892941192f90aadd4f452a295bf39fc5f7ed 229 textDocument/hover 3ms
2017-01-19T20:53:55.282136382Z <<< git://github.com/uber/tchannel-go?9a5445fa1ae4d24ce3407eaf7069d13a0a63deac 0 initialize 1110ms
2017-01-19T20:53:55.287035875Z >>> git://github.com/uber/tchannel-go?9a5445fa1ae4d24ce3407eaf7069d13a0a63deac 1 textDocument/hover {"position":{"character":0,"line":0},"textDocument":{"uri":"file:///fragmenting_reader.go"}}
2017-01-19T20:53:55.470055820Z >>> git://github.com/captncraig/mux?acfc892941192f90aadd4f452a295bf39fc5f7ed 230 textDocument/definition {"position":{"character":5,"line":23},"textDocument":{"uri":"file:///mux.go"}}
2017-01-19T20:53:55.472086920Z <<< git://github.com/captncraig/mux?acfc892941192f90aadd4f452a295bf39fc5f7ed 230 textDocument/definition 2ms
2017-01-19T20:53:56.368889652Z <<< git://github.com/uber/tchannel-go?9a5445fa1ae4d24ce3407eaf7069d13a0a63deac 1 textDocument/hover 1081ms
2017-01-19T20:54:18.271657581Z >>> git://github.com/openzipkin/zipkin-go-opentracing?594640b9ef7e5c994e8d9499359d693c032d738c 0 initialize {"rootPath":"file:///","rootUri":"file:///","capabilities":{"xfilesProvider":true,"xcontentProvider":true},"originalRootURI":"git://github.com/openzipkin/zipkin-go-opentracing?594640b9ef7e5c994e8d9499359d693c032d738c","mode":"go"}
2017-01-19T20:54:18.399144361Z <<< git://github.com/openzipkin/zipkin-go-opentracing?594640b9ef7e5c994e8d9499359d693c032d738c 0 initialize 127ms
2017-01-19T20:54:18.400137560Z >>> git://github.com/openzipkin/zipkin-go-opentracing?594640b9ef7e5c994e8d9499359d693c032d738c 1 textDocument/hover {"position":{"character":0,"line":0},"textDocument":{"uri":"file:///examples/middleware/http.go"}}'''.splitlines())
