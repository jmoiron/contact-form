# contact-form

A simple stand-alone web server which implements a contact form on an
otherwise static website.

## usage

If you run `contact-form` in a directory, it will serve up everything in that
directory as a static file in the same way a server like `nginx` or `apache`
would, defaulting to `index.html` when a "directory" is accessed.

It also sets up a listener on `/contact/`, which does a few things:

* You can `POST` with `email`, `subject`, and `message` to attempt to send
  an email.  This returns a json message `{"success": true}` on success and
  `{"success": false, "error": "<an error message"}` on failure.

* If there are validation erorrs on particular fields (eg message length,
  invalid email address, etc.), then those fields will have an error message
  in the returned json, eg: `{"email": "invalid email address"}`.

## configuring


