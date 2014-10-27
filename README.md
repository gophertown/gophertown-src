gophertown.org
==============

This repository is the source code for
[gophertown.org](http://gophertown.org). To add or modify your data on the
website, you want the [gophertown-data repo](https://github.com/gophertown/gophertown-data).

The `gtown` directory contains the API backend.  It loads a directory of json
files and serves them up to the front-end.

The `gophertown.org` directory contains the frontend; `index.html` is the entry
point. This is an ember.js app and is entirely client-based. It uses the
backend API to get data.

To get this running:

```
bash$ go get github.com/gophertown/gophertown-src/gtown
bash$ git clone https://github.com/gophertown/gophertown-data.git
bash$ gtown -gopherdir=gophertown-data/data/ -site=$GOPATH/src/github.com/gophertown/gophertown-src/gophertown.org/
```

## License

The backend code is copyright 2014 gophertown.org developers.

The client-side code was originally from rustaceans.org, but has been modified
to make it more appropriate for displaying gophers.  It has the following
license:

copyright 2014 rustaceans.org developers.

Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
<LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
option. This file may not be copied, modified, or distributed
except according to those terms.
