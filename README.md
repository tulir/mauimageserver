# mauImageServer
![Build Status](https://git.maunium.net/Tulir293/mis2/badges/master/build.svg)

## Introduction
mauImageServer is a simple image hosting and sharing backend designed to be used with [mauCapture](https://git.maunium.net/Tulir293/maucapture2). It uses [mAuth](https://git.maunium.net/Tulir293/mauth) for authentication. It has a basic search function

## API
### Authentication
The login interface is located at `/auth/login` and register at `/auth/register`. See the documentation of [mAuth](https://git.maunium.net/Tulir293/mauth) for details about the request payload.

### Requests
Interfaces:
* `/insert` - Insert images
* `/delete` - Delete images (requires authentication)
* `/search` - Search for images
* `/hide` - Hide images from search (requires authentication)

TODO: Better documentation

### Responses
TODO: Documentation
