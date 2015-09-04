#Relay
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/relay)

Relay is a simple microframework with very simple designs that provide you with the minimal tools you need to get things running.

# Download

     go get github.com/influx6/relay

# Install

     go install github.com/influx6/relay 


#Features

  - Codecs: Relay builds on the ideas that response should be dynamic and customizable,rather than creating a writer, response feed like approach, Relay uses codecs that take a response type or format and perform the operation of converting the response into the appropriate response object. This allows a custom message pattern to be created eg. binary only messages on top of http or websocket connections. Codecs have the pairs of encoders and decoders that provide the basis of transmission and reception.

  - Chainable Ports: Ports are just encapsulated, chain handlers of request which is a standard feature of most golang middleware based framework.

  - Controllers: this provide a basic encapsulation principle for a group of routine handlers and not in the sense of c of mvc

#Example

#License

  . [MIT]()
