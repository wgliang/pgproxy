![pgproxy](./pgproxy.png)

# pgproxy
[![Build Status](https://travis-ci.org/wgliang/pgproxy.svg?branch=master)](https://travis-ci.org/wgliang/pgproxy)
[![codecov](https://codecov.io/gh/wgliang/pgproxy/branch/master/graph/badge.svg)](https://codecov.io/gh/wgliang/pgproxy)
[![GoDoc](https://godoc.org/github.com/wgliang/pgproxy?status.svg)](https://godoc.org/github.com/wgliang/pgproxy)
[![Code Health](https://landscape.io/github/wgliang/pgproxy/master/landscape.svg?style=flat)](https://landscape.io/github/wgliang/pgproxy/master)
[![Code Issues](https://www.quantifiedcode.com/api/v1/project/98b2cb0efd774c5fa8f9299c4f96a8c5/badge.svg)](https://www.quantifiedcode.com/app/project/98b2cb0efd774c5fa8f9299c4f96a8c5)
[![Go Report Card](https://goreportcard.com/badge/github.com/wgliang/pgproxy)](https://goreportcard.com/report/github.com/wgliang/pgproxy)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)

pgproxy is a postgresql proxy server, through a pipe redirect connection,then you can filter the requested sql statement. The future it will support multi-database backup, adapt to distributed databases and other scenes except analysis sql statement.


## Installation

```
$ go get -u github.com/wgliang/pgproxy
```

## Using

Start or shut down the proxy server.
```
$ pgproxy start/stop
```

Use pgproxy on the command line
```
$ pgproxy cli
```

Ps: You can use it as you would with a native command line.

## Support

select/delete/update statement and support any case.

## Credits

package parser is base on [sqlparser](https://github.com/xwb1989/sqlparser)