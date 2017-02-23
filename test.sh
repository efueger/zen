#! /bin/bash

echo 'Run test and benchmark, race detect and test coverage'
go test -v -bench . -behchmem -race -cover

