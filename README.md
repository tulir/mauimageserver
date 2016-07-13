# mauImageServer
[![License](http://img.shields.io/:license-gpl3-blue.svg?style=flat-square)](http://www.gnu.org/licenses/gpl-3.0.html)

mauImageServer is a simple image hosting and sharing backend designed to be used with [mauCapture](https://github.com/tulir293/maucapture2).
[mAuth](https://github.com/tulir293/mauth) is used for authentication.
It also has a basic search function. An example search frontend can be found from [img.mau.lu/search.html](https://img.mau.lu/search.html).

## Setup
### Packaging & Install
You can generate a debian package using `make package`. It will produce a deb package named `mauimageserver.deb`.

You can use `sudo dpkg -i mauimageserver.deb` to install the package.

### Configuration
* `image-location` - The location to store uploaded images
* `date-format` - The Go date format to display when using the image template
* `require-auth` - Require authentication (mAuth) to upload images. Removing/Hiding/Replacing images always requires authentication
* `trust-headers` - Trust the `X-Forwarded-For` header usually set by load balancers or using proxy pass in a web server
* `allow-search` - Allow searching for images based on various factors
* `sql` - MySQL/MariaDB settings
  * `connection` - Connection details
    * `mode` - The mode to connect using (Usually `tcp` or `unix`)
    * `ip` - The IP or Unix socket path to connect to
    * `port` - The port to connect to (when using TCP)
  * `authentication` - Datbase authentication. Should be self-explanatory
  * `database` - The name of the database to use. The database must exist, but tables will be created automatically

Default configuration:
```json
{
    "image-location": "/var/mis",
    "date-format": "15:04:05 02.01.2006 MST",
    "image-template": "/etc/mis/image.html",
    "require-auth": true,
    "trust-headers": false,
    "allow-search": true,
    "ip": "127.0.0.1",
    "port": 29300,
    "sql": {
        "connection": {
            "mode": "tcp",
            "ip": "127.0.0.1",
            "port": 3306
        },
        "authentication": {
            "username": "root",
            "password": "password"
        },
        "database": "mauimageserver"
    }
}
```

## API
### Authentication
The login interface is located at `/auth/login` and register at `/auth/register`. See the documentation of [mAuth](https://github.com/tulir293/mauth) for details about the request payload.

### Requests
#### Insert
An insert request can have the following fields:
 * `image` - The image file encoded in base64. **Required for all insert requests**
 * `image-name` - The requested image name. If the image name is already used by someone else, this will return the error `already-exists`. If the image name is used by the person uploading a new image, it will be replaced and the status will be `replaced` instead of `created`.
 * `image-format` - The image name extension. This is just for the direct URL as the MIME type will be determined from the image itself.
 * `client-name` - The name of the client used to upload the image. This is purely for statistics and search.
 * `username` - Username for authentication.
 * `auth-token` - Authentication token.
 * `hidden` - Whether or not to hide the image automatically.

#### Delete
A delete request requires authentication and the image being deleted must obviously be uploaded by the user trying to delete the image.

A delete request must have the following fields:
 * `image-name` - The name of the image to be deleted.
 * `username` - Username for authentication.
 * `auth-token` - Authentication token.

#### Hide
A hide request is similar to a delete request. It too requires authentication and the user trying to hide the image must be the one who uploaded it.

In addition to the fields of a delete request, a hide request must also have the field `hidden` which must be a boolean value of whether or not the image should be hidden.

#### Search
A search query may contain the following fields:
 * `image-format` - The format of the image.
 * `uploader` - The username of the person who uploaded the image. Doesn't have to be exactly correct, a part of the username should work.
 * `client-name` - The client used to upload the image. As with the uploader, doesn't have to be exact.
 * `uploaded-after` - Only include images uploaded after this unix timestamp.
 * `uploaded-before` - Only include images uploaded before this unix timestamp.
 * `auth-token` - Authentication token. Must be used with exact username in the `uploader` field. When used, hidden images will be returned.

### Responses
Insert, Delete and Hide requests will respond with the same JSON template, which contains the following fields:
 * `success` - Whether or not the action was successful.
 * `status-simple` - A simple and short error keyword.
 * `status-humanreadable` - A longer, human-readable error message.

A search query will respond with the same JSON template as the other requests, but in addition to that there will be an array of search results, which will contain the following fields:
 * `image-name` - The name of the image. Does not contain the extension (see `image-format`)
 * `image-format` - The file name extension of the image.
 * `mime-type` - The MIME type of the image.
 * `adder` - The name of the user who uploaded the image.
 * `client-name` - The name of the client used to upload the image.
 * `timestamp` - The unix timestamp of the time the image was uploaded.
 * `id` - The index of the image. Indexes start from 0 and increment by one for each image uploaded.
 * `hidden` - Whether or not the image is hidden from non-authenticated search.
