Jekyll Static Site Server
=========================

My first Golang project. Inspired by [Hacking with Andrew and Brad - tip.golang.org](https://www.youtube.com/watch?v=1rZ-JorHJEY),
this GO fileserver serves my site from two folders; "BUILD_A" and "BUILD_B".
Using Git (or bitbucket) webhooks to trigger the git update process,
it checks the hash of the latest commits to determine when to rebuild the site.

Serve from BUILD A -> New Post -> Webhooks -> Build into BUILD_B -> Serve site from BUILD_B

## Usage

You may want to modify some of the code on how it updates to fit your needs. I separated my Jekyll site and posts into
two different repos; you may want to have everything in one repo.

To build, use the included Dockerfile and run `docker build .` in the directory.

## License

The MIT License (MIT)

Copyright (c) 2015 Vincent Cheung.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit
persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
