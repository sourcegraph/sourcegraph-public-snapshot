from twisted.internet.protocol import Protocol, ReconnectingClientFactory
from twisted.internet import task
from sys import stdout

import encode

class CollectorProtocol(Protocol):
    # CollectorProtocol is a Twisted implementation of appdash's protobuf-based
    # collector protocol.

    # writeMsg writes the delimited-protobuf message out to the protocol's
    # transport. See encodeMsg for details.
    def writeMsg(self, msg):
        self.transport.write(encode._msg(msg))

    def connectionMade(self):
        self._factory._log('connected!')

    def connectionLost(self, reason):
        self._factory._log("disconnected.", reason.getErrorMessage())

    def dataReceived(self, data):
        self._factory._log('got', len(data), 'bytes of unexpected data from server.')


class RemoteCollectorFactory(ReconnectingClientFactory):
    # RemoteCollectorFactory is a Twisted factory for remote collectors, which
    # collect spans and their annotations, sending them to a remote Go appdash
    # server for collection. After collection they can be viewed in appdash's
    # web user interface.

    _reactor = None
    _debug = False
    _remote = None
    _pending = []

    def __init__(self, reactor, debug=False):
        self._reactor = reactor
        self._debug = debug

    def _log(self, *args):
        if self._debug:
            print "appdash: %s" % (" ".join(args))

    # collect collects annotations for the given spanID.
    #
    # The annotations will be flushed out at a later time, when a connection
    # to the remote server has been made.
    def collect(self, spanID, *annotations):
        self._log("collecting", str(len(annotations)), "annotations for", str(spanID))

        # Append the collection packet to the pending queue.
        self._pending.append(encode._collect(spanID, *annotations))

    # __flush is called internally after either a new collection has occured, or
    # after connection has been made with the remote server. It writes all the
    # pending messages out to the remote.
    def __flush(self):
        if len(self._pending) == 0:
            return

        self._log("flushing", str(len(self._pending)), "messages")
        for p in self._pending:
            self._remote.writeMsg(p)
        self._log("done.")
        self._pending = []

    def __startFlushing(self):
        # Run the flush method every 1/2 second.
        l = task.LoopingCall(self.__flush)
        l.start(1/2)

    def startedConnecting(self, connector):
        self._log('connecting..')

    def buildProtocol(self, addr):
        # Reset delay to reconnection -- otherwise it's exponential (which is
        # not a good match for us).
        self.resetDelay()

        # Create the protocol.
        p = CollectorProtocol()
        p._factory = self
        self._remote = p
        self._reactor.callLater(1, self.__startFlushing)
        return p

    def clientConnectionFailed(self, connector, reason):
        self._log('connection failed:', reason.getErrorMessage())
        ReconnectingClientFactory.clientConnectionFailed(self, connector, reason)

