.PHONY:	test golint dep

# Copyright 2015 MediaMath <http://www.mediamath.com>.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

ifeq ($(CIRCLECI),true)
  TEST_RACE=
else
  TEST_RACE=-race
endif

test: golint dep
	golint ./...
	go vet ./...
	go test $(VERBOSITY) $(TEST_RACE) ./...

golint:
	go get github.com/golang/lint/golint

dep: 
	go get gopkg.in/stretchr/testify.v1
	go get ./...

