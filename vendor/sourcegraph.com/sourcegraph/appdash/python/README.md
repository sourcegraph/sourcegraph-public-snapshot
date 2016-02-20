# AppDash: Python

Appdash-python enables Python-based applications to send performance and debug information to a remote Go-based Appdash collection server.

The `appdash` module provides everything that is needed to get started, as well as two collectors (an asynchronous one based on the Twisted networking library, and a synchronous one Python's socket module).

## Integration

At this point, there is no integration with common Python web-frameworks -- although adding support for such should be trivial.

Generally speaking if the web framework has some form of integration with the Twisted networking library, using the Twisted collector (`appdash.twcollector`) is a better choice as it is asynchronous and Twisted is better at scheduling tasks.

Using Twisted is not always possible, for example Django applications that are hosted by Apache (and thus, Twisted cannot be long-running). To suit these cases, a synchronous remote collector is provided (`appdash.sockcollector`) which operates using Python's standard socket module.

For quick'n'dirty integration, the synchronous collector is probably more straight-forward as well (i.e. if you are not accustomed to working with Twisted's asynchronous programming model).

## Prerequisites

To install appdash for python you'll first need to install a few things through the standard `easy_install` and `pip` python package managers:

```
# (Ubuntu/Linux) Install easy_install and pip:
sudo apt-get install python-setuptools python-pip

# Install Google's protobuf:
easy_install protobuf

# Install Twisted networking (optional, only needed for Twisted integration):
easy_install twisted

# Install strict-rfc3339
pip install strict-rfc3339
```

Depending on where Python is installed and/or what permissions the directory has, you may need to run the above `easy_install` and `pip` commands as root.

## Installation

To install Appdash into your Python path (i.e. so you can import it into your own code), simply change directory to `sourcegraph.com/sourcewgraph/appdash/python` and run the traditional `setup.py` script:

```
# Install Python-appdash:
cd $GOPATH/src/sourcegraph.com/sourcegraph/appdash/python
python setup.py install
```

Again, depending on where Python is installed and/or what permissions the directory has, you may need to run the above `setup.py` script as root.

To test that installation went well, in any directory _except that one_, you can launch an interactive Python interpreter and simply `import appdash`.

## Twisted Example

If all is well with your setup, you should be able to change directory to `sourcegraph.com/sourcewgraph/appdash/python` and run the Twisted example:

```
# Run appdash server in separate terminal:
appdash serve

# Run the example script:
cd $GOPATH/src/sourcegraph.com/sourcegraph/appdash/python
./example_twisted.py
```

## Socket Example

If you prefer not to use Twisted, you can utilize the standard socket collector (`appdash.sockcollector`) provided. You can run the example:

```
# Run appdash server in separate terminal:
appdash serve

# Run the example script:
cd $GOPATH/src/sourcegraph.com/sourcegraph/appdash/python
./example_socket.py
```
