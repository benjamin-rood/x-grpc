# Golang Homework Assignment

You are expected to create a simple REST/GRPC server application, publish it to a public git repository on a service like GitHub.

## Basic Requirements

* The server application written in Go Programming Language
* The server should serve a single REST endpoint that allows to upload a single file
* It should be possible to upload the whole content of a file of unlimited size in a single REST request
* It is preferred that only a limited size chunk of transferred data is stored in memory and then written back to a disk file
* The automated test cases that demonstrate that the file upload can be initiated from the client side should be added to the same git repository as unit tests
* It is expected that the test case uses client side connection to connect over GRPC to the server
* It is not expected that the client side application to connect to the server is provided though

## Bonus Requirements for Extra Credit

* Instead of REST endpoint the server serves GRPC (https://grpc.io/) with the following requirements:
  - The server provides one GRPC service with a single method that allows to upload a single file
  - The whole content of a file can be uploaded in a single method call
  - It is expected that the test case uses client side connection to connect over GRPC to the server
* After the content is written to a file on the server the attempt is made to parse it as JSON data
* If the JSON unmarshalling succeeds then the following modifications are done in the JSON data and the result is written out (marshaled) to a file with a different name:
  - The properties that start with a vowel should be removed from the JSON data
  - The properties that have even integer number should be increased by *1000*
  - It is expected that the corresponding automated test coverage is included

## Evaluation Criteria

* Your result will be evaluated based on how well all requirements are implemented.