// Copyright 2013-2019 Laurent Moussault <laurent.moussault@gmail.com>
// SPDX-License-Identifier: BSD-2-Clause

package loop

////////////////////////////////////////////////////////////////////////////////

type private struct{}

// An Option represents a configuration option used to change some parameters of
// the loop: see Configure.
type Option func(*private) error

var options []Option

// Configure the loop.
func Configure(o ...Option) {
	options = append(options, o...)
}
