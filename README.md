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

## spam control

By default, `contact-form` will check incoming messages against the
[postmark spamcheck API](http://spamcheck.postmarkapp.com/doc).  If you do
not want this, or it gets discontinued or changed, you can disable this with
the `-nospam` flag or an environ variable.

## configuring

You can use the following environment variables to control the way that
`contact-form` behaves and sends email.

```
CONTACT_PORT        listener http port
CONTACT_NOSPAM      if set, do not check messages for spam
CONTACT_MAILPORT    outgoing mail server port
CONTACT_MAILHOST    outgoing mail server host
CONTACT_MAILUSER    outgoing mail server username
CONTACT_MAILPASS    outgoing mail server passsword
CONTACT_DESTEMAIL   destination email address for messages
```

