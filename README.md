Jekyll Static Site Server
=========================

My first Golang project. Inspired by [Hacking with Andrew and Brad - tip.golang.org](https://www.youtube.com/watch?v=1rZ-JorHJEY),
this Go fileserver serves my site from two folders: "BUILD_A" and "BUILD_B".
When I author a new post, Git webhooks is used to trigger the update routine. The routine checks for new commit hashes and rebuilds the site if new hashes were observed.

It works like this:

Build/Serve site from BUILD A -> New Post -> (git push) -> Webhooks -> Build into BUILD_B -> Serve site from BUILD_B

## Usage

You may want to modify some of the code on how it updates to fit your needs. I separated my Jekyll site and posts into two different repos; you may want to have everything in one repo.

Change the repo strings in the `defaults.yml` file. Add your credentials in a `settings.yml` file (which overrides defaults).

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
