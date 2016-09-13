package sr

//Copyright 2016 MediaMath <http://www.mediamath.com>.  All rights reserved.
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

import "fmt"

//Subject is *not* a topic.  Subject is a schema registry abstraction, topic is a kafka one.  For instance the kafka rest proxy assumes that for
//topic "foo" there can be 2 subjects foo-key and foo-value.  foo-key will store the schema for the key field (if any) and foo-value will store the
//schema for the value field
type Subject string

//EmptySubject is just a place holder for the empty string.
const EmptySubject = Subject("")

//ValueSubject takes a topic name and turns it into what kafka-rest assumes is the name for value schemas
func ValueSubject(topic string) Subject {
	if topic == "" {
		return EmptySubject
	}

	return Subject(fmt.Sprintf("%s-value", topic))
}

//KeySubject takes a topic name and turns it into what kafka-rest assumes is the name for value schemas
func KeySubject(topic string) Subject {
	if topic == "" {
		return EmptySubject
	}

	return Subject(fmt.Sprintf("%s-key", topic))
}
