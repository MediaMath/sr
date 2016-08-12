.PHONY:	test golint dep

# Copyright 2015 MediaMath <http://www.mediamath.com>.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

ifdef VERBOSE
	TEST_VERBOSITY=-v
else
	TEST_VERBOSITY=
endif

test: golint dep
	golint ./...
	go vet ./...
	go test $(TEST_VERBOSITY) -race ./...

golint:
	go get github.com/golang/lint/golint

dep: 
	go get gopkg.in/stretchr/testify.v1
	go get ./...

