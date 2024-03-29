// mgo - MongoDB driver for Go
// 
// Copyright (c) 2010-2011 - Gustavo Niemeyer <gustavo@niemeyer.net>
// 
// All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 
//     * Redistributions of source code must retain the above copyright notice,
//       this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright notice,
//       this list of conditions and the following disclaimer in the documentation
//       and/or other materials provided with the distribution.
//     * Neither the name of the copyright holder nor the names of its
//       contributors may be used to endorse or promote products derived from
//       this software without specific prior written permission.
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

import (
	"launchpad.net/gocheck"
)

type QS struct{}

var _ = gocheck.Suite(&QS{})

func (s *QS) TestSequentialGrowth(c *gocheck.C) {
	q := queue{}
	n := 2048
	for i := 0; i != n; i++ {
		q.Push(i)
	}
	for i := 0; i != n; i++ {
		c.Assert(q.Pop(), gocheck.Equals, i)
	}
}

var queueTestLists = [][]int{
	// {0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},

	// {8, 9, 10, 11, ... 2, 3, 4, 5, 6, 7}
	{0, 1, 2, 3, 4, 5, 6, 7, -1, -1, 8, 9, 10, 11},

	// {8, 9, 10, 11, ... 2, 3, 4, 5, 6, 7}
	{0, 1, 2, 3, -1, -1, 4, 5, 6, 7, 8, 9, 10, 11},

	// {0, 1, 2, 3, 4, 5, 6, 7, 8}
	{0, 1, 2, 3, 4, 5, 6, 7, 8,
		-1, -1, -1, -1, -1, -1, -1, -1, -1,
		0, 1, 2, 3, 4, 5, 6, 7, 8},
}


func (s *QS) TestQueueTestLists(c *gocheck.C) {
	test := []int{}
	testi := 0
	reset := func() {
		test = test[0:0]
		testi = 0
	}
	push := func(i int) {
		test = append(test, i)
	}
	pop := func() (i int) {
		if testi == len(test) {
			return -1
		}
		i = test[testi]
		testi++
		return
	}

	for _, list := range queueTestLists {
		reset()
		q := queue{}
		for _, n := range list {
			if n == -1 {
				c.Assert(q.Pop(), gocheck.Equals, pop(),
					gocheck.Bug("With list %#v", list))
			} else {
				q.Push(n)
				push(n)
			}
		}

		for n := pop(); n != -1; n = pop() {
			c.Assert(q.Pop(), gocheck.Equals, n,
				gocheck.Bug("With list %#v", list))
		}

		c.Assert(q.Pop(), gocheck.Equals, nil,
			gocheck.Bug("With list %#v", list))
	}
}
