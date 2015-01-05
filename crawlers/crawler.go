// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crawlers provides crawlers for gathering source code repository
// information.
package crawlers

// Crawler defines methods a crawler must implement.
type Crawler interface {
	// Crawl methods crawls data and put it into the database.
	Crawl()
}
